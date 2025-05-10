// Package patcher is a library to handle modification of uGO Bytecode.

package patcher

import (
	"fmt"
	"runtime"
	"sync"
	"time"

	"github.com/ozanh/ugo"
)

const (
	patchNext byte = iota
	patchInsertBefore
)

type patchFunc = func(*instsIterator) (op byte, insts []byte)

// PatchForGosched modifies given ugo.Bytecode to add a callable to the given
// ugo.Bytecode that tries to park the VM goroutine when the number of calls to
// the callable reaches the given threshold. This patch should be used in single
// threaded application e.g. WebAssembly. If error is returned, given
// ugo.Bytecode must be discarded due to invalid patching.
func PatchForGosched(bc *ugo.Bytecode, callThreshold uint32) (int, error) {
	// Generate following instructions to insert before backward jumps and
	// function start points.
	/*
		0000 CONSTANT <index>
		0000 CALL 0 0
		0000 POP
	*/

	if callThreshold == 0 {
		panic("callThreshold must be greater than 0")
	}

	constIndex := len(bc.Constants)
	insert := make([]byte, 0, 7)
	b := make([]byte, 8)
	b, err := ugo.MakeInstruction(b, ugo.OpConstant, constIndex)
	if err != nil {
		return 0, err
	}
	insert = append(insert, b...)
	b, err = ugo.MakeInstruction(b, ugo.OpCall, 0, 0)
	if err != nil {
		return 0, err
	}
	insert = append(insert, b...)
	b, err = ugo.MakeInstruction(b, ugo.OpPop)
	if err != nil {
		return 0, err
	}
	insert = append(insert, b...)
	var numInserts int
	bp := newBytecodePatcher(bc,
		func(it *instsIterator) (byte, []byte) {
			pos := it.Pos()
			if pos == 0 {
				// insert at the top of function
				numInserts++
				return patchInsertBefore, insert
			}
			opcode := it.Opcode()
			if opcode == ugo.OpJump {
				// if jump backward, insert instructions before jump
				if it.Operands()[0] < pos {
					numInserts++
					return patchInsertBefore, insert
				}
			}
			return patchNext, nil
		},
	)
	if err := bp.patch(); err != nil {
		return numInserts, err
	}

	fn := &goschedFunc{
		callThreshold: callThreshold,
		sleep:         runtime.NumCPU() == 1 || runtime.GOMAXPROCS(0) == 1,
	}
	bc.Constants = append(bc.Constants, fn)
	return numInserts, nil
}

type goschedFunc struct {
	ugo.ObjectImpl
	mu            sync.Mutex
	numCalls      uint64
	counter       uint32
	callThreshold uint32
	sleep         bool
}

var _ ugo.ExCallerObject = (*goschedFunc)(nil)

func (g *goschedFunc) String() string   { return "<gosched>" }
func (g *goschedFunc) TypeName() string { return g.String() }
func (g *goschedFunc) CanCall() bool    { return true }

func (g *goschedFunc) Call(args ...ugo.Object) (ugo.Object, error) {
	return g.CallEx(ugo.Call{})
}

func (g *goschedFunc) CallEx(_ ugo.Call) (ugo.Object, error) {
	g.mu.Lock()
	defer g.mu.Unlock()

	g.numCalls++
	g.counter++
	if g.counter == g.callThreshold {
		g.counter = 0

		runtime.Gosched()

		if g.sleep {
			//lint:ignore SA1004 // Park the current goroutine.
			time.Sleep(1) // I couldn't find another way to park the goroutine.
		}
	}
	return ugo.Undefined, nil
}

type bytecodePatcher struct {
	it       *instsIterator
	bc       *ugo.Bytecode
	jumps    []posJump
	smap     sourceMapper
	newInsts []byte
	curInsts []byte
	modifier patchFunc
}

func newBytecodePatcher(bc *ugo.Bytecode, fn patchFunc) *bytecodePatcher {
	bm := &bytecodePatcher{
		bc:       bc,
		it:       &instsIterator{operands: make([]int, 4)},
		modifier: fn,
	}
	return bm
}

func (bp *bytecodePatcher) patch() (err error) {
	curFn := bp.bc.Main
	numConsts := len(bp.bc.Constants)
	cidx := -1
	for cidx < numConsts {
		bp.curInsts = curFn.Instructions
		bp.newInsts = make([]byte, 0, cap(bp.curInsts))
		bp.smap.Reset(curFn.SourceMap)
		if err = bp.saveJumpPos(); err != nil {
			return
		}
		if err = bp.generate(); err != nil {
			return
		}
		if err = bp.updateJumps(); err != nil {
			return
		}
		curFn.Instructions = bp.newInsts
		curFn.SourceMap = bp.smap.MakeSourceMap()

		cidx++
		for cidx < numConsts {
			if f, ok := bp.bc.Constants[cidx].(*ugo.CompiledFunction); ok {
				curFn = f
				break
			}
			cidx++
		}
	}
	return
}

func (bp *bytecodePatcher) saveJumpPos() error {
	bp.jumps = bp.jumps[:0]
	bp.it.Reset(bp.curInsts)
	for bp.it.Next() {
		switch op := bp.it.Opcode(); op {
		case ugo.OpJumpFalsy,
			ugo.OpJump,
			ugo.OpAndJump,
			ugo.OpOrJump,
			ugo.OpSetupTry:
			pos := bp.it.Pos()
			operands := bp.it.Operands()
			bp.jumps = append(bp.jumps,
				posJump{
					pos:     pos,
					jump:    operands[0],
					opcode:  op,
					operand: 0,
				},
			)
			if op == ugo.OpSetupTry {
				bp.jumps = append(bp.jumps,
					posJump{
						pos:     pos,
						jump:    operands[1],
						opcode:  op,
						operand: 1,
					},
				)
			}
		}
	}
	return bp.it.Error()
}

func (bp *bytecodePatcher) generate() error {
	bp.it.Reset(bp.curInsts)
	for bp.it.Next() {
		op, insts := bp.modifier(bp.it)
		pos, offset := bp.it.Pos(), bp.it.Offset()
		switch op {
		case patchNext:
			bp.newInsts = append(bp.newInsts, bp.curInsts[pos:pos+offset+1]...)
		case patchInsertBefore:
			bp.insertAt(len(bp.newInsts), len(insts))
			bp.newInsts = append(bp.newInsts, insts...)
			bp.newInsts = append(bp.newInsts, bp.curInsts[pos:pos+offset+1]...)
		default:
			return fmt.Errorf("generate: unknown op: %d", op)
		}
	}
	return bp.it.Error()
}

func (bp *bytecodePatcher) insertAt(pos, size int) {
	for i := 0; i < len(bp.jumps); i++ {
		bp.jumps[i].InsertAt(pos, size)
	}
	bp.smap.InsertAt(pos, size)
}

func (bp *bytecodePatcher) updateJumps() error {
	operands := make([]int, 0, 2)
	for _, v := range bp.jumps {
		if !v.updated {
			continue
		}
		if bp.newInsts[v.pos] != v.opcode {
			msg := "updateJumps: opcodes expected: %d, got: %d"
			return fmt.Errorf(msg, v.opcode, bp.newInsts[v.pos])
		}
		operands, _ = ugo.ReadOperands(
			ugo.OpcodeOperands[v.opcode],
			bp.newInsts[v.pos+1:],
			operands,
		)
		operands[v.operand] = v.jump
		insts := make([]byte, 8)
		insts, err := ugo.MakeInstruction(insts, v.opcode, operands...)
		if err != nil {
			return fmt.Errorf("updateJumps: %w", err)
		}
		copy(bp.newInsts[v.pos:], insts)
	}
	return nil
}

// posJump holds the jump instructions data to be able to update position and
// jump target positions when instructions are inserted.
type posJump struct {
	pos     int
	jump    int
	opcode  byte
	operand byte
	updated bool
}

func (pj *posJump) InsertAt(pos, size int) {
	if pj.pos >= pos {
		pj.updated = true
		pj.pos += size
	}
	if pj.jump < pos || (pj.opcode == ugo.OpSetupTry && pj.jump == 0) {
		return
	}
	pj.updated = true
	pj.jump += size
}

// instsIterator is a lazy instructions iterator that gets operands on demand.
// Use Reset method to re-use the same instance.
type instsIterator struct {
	pos      int
	insts    []byte
	opcode   ugo.Opcode
	operands []int
	offset   int
	err      error
}

func (it *instsIterator) Next() bool {
	if it.pos >= len(it.insts) || it.err != nil {
		return false
	}
	it.opcode = it.insts[it.pos]
	if int(it.opcode) >= len(ugo.OpcodeOperands) {
		it.err = fmt.Errorf("invalid opcode %d at %d", it.opcode, it.pos)
		return false
	}
	it.offset = opWidths[it.opcode]
	it.pos += it.offset + 1
	return true
}

func (it *instsIterator) Opcode() ugo.Opcode {
	return it.opcode
}

// Returning slice is reused at next call, copy if required.
func (it *instsIterator) Operands() []int {
	it.operands, _ = ugo.ReadOperands(
		ugo.OpcodeOperands[it.opcode],
		it.insts[it.pos-it.offset:],
		it.operands,
	)
	return it.operands
}

func (it *instsIterator) Offset() int {
	return it.offset
}

func (it *instsIterator) Pos() int {
	return it.pos - it.offset - 1
}

func (it *instsIterator) Error() error {
	return it.err
}

func (it *instsIterator) Reset(insts []byte) {
	it.pos = 0
	it.insts = insts
	it.err = nil
}

// sourceMapper holds source map of a ugo.CompiledFunction to update instruction
// positions in map by converting the source map to two slices as keys and values
// and updating keys (positions) at every insertion.
type sourceMapper struct {
	keys   []int
	values []int
}

func (sm *sourceMapper) Reset(sourceMap map[int]int) {
	if sm.keys == nil {
		sm.keys = make([]int, len(sourceMap))
		sm.values = make([]int, len(sourceMap))
	}
	sm.keys = sm.keys[:0]
	sm.values = sm.values[:0]
	for k, v := range sourceMap {
		sm.keys = append(sm.keys, k)
		sm.values = append(sm.values, v)
	}
}

func (sm *sourceMapper) InsertAt(pos, size int) {
	// Sequential search is mostly 30% faster due to sort overhead for small slices.
	for i, v := range sm.keys {
		if v >= pos {
			sm.keys[i] = v + size
		}
	}
}

func (sm *sourceMapper) MakeSourceMap() map[int]int {
	m := make(map[int]int, len(sm.keys))
	for i, v := range sm.keys {
		m[v] = sm.values[i]
	}
	return m
}

var opWidths = getOpWidths()

func getOpWidths() (opWidths []int) {
	opWidths = make([]int, len(ugo.OpcodeOperands))
	for op := range ugo.OpcodeOperands {
		for _, w := range ugo.OpcodeOperands[op] {
			opWidths[op] += w
		}
	}
	return opWidths
}
