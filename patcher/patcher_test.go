package patcher_test

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ozanh/ugodev/patcher"

	. "github.com/ozanh/ugo"
)

func TestGosched(t *testing.T) {
	opts := CompilerOptions{}
	expectCompile(t, ``, opts, func(bc *Bytecode) {
		expected := copyBytecode(bc)
		n, err := patcher.Gosched(expected, 100)
		require.NoError(t, err)
		require.Equal(t, 1, n)
		expected.Constants = expected.Constants[:len(expected.Constants)-1]
		expectPatch(t, bc, expectedPatch{
			bc:         expected,
			numCalls:   1,
			numInserts: 1,
		})
	})

	expectCompile(t, `for { }`, opts, func(bc *Bytecode) {
		expected := copyBytecode(bc)
		expectPatch(t, bc, expectedPatch{
			bc:            expected,
			noCmpCompFunc: true,
			noCmpConsts:   true,
			noExec:        true,
			numInserts:    1,
		})
	})
	expectCompile(t, `for { 1; }`, opts, func(bc *Bytecode) {
		expected := copyBytecode(bc)
		expectPatch(t, bc, expectedPatch{
			bc:            expected,
			noCmpCompFunc: true,
			noCmpConsts:   true,
			noExec:        true,
			numInserts:    2,
		})
	})
	expectCompile(t, `for { 1; }; for { 1;}`, opts, func(bc *Bytecode) {
		expected := copyBytecode(bc)
		expectPatch(t, bc, expectedPatch{
			bc:            expected,
			noCmpCompFunc: true,
			noCmpConsts:   true,
			noExec:        true,
			numInserts:    3,
		})
	})
	expectCompile(t, `for i := 0; i < 9; i++ { }`, opts, func(bc *Bytecode) {
		expected := copyBytecode(bc)
		expectPatch(t, bc, expectedPatch{
			bc:            expected,
			noCmpCompFunc: true,
			noCmpConsts:   true,
			numInserts:    2,
			callCount:     11,
			numCalls:      10,
		})
	})
	expectCompile(t, `for i := 0; i < 9; i++ { }`, opts, func(bc *Bytecode) {
		expected := copyBytecode(bc)
		expectPatch(t, bc, expectedPatch{
			bc:            expected,
			noCmpCompFunc: true,
			noCmpConsts:   true,
			numInserts:    2,
			callCount:     9,
			numCalls:      1,
		})
	})
	expectCompile(t, `f := func() {}; f()`, opts, func(bc *Bytecode) {
		expected := copyBytecode(bc)
		expectPatch(t, bc, expectedPatch{
			bc:            expected,
			noCmpCompFunc: true,
			noCmpConsts:   true,
			numInserts:    2,
			numCalls:      2,
		})
	})
	expectCompile(t, `
	var fib
	fib = func(x) {
		return x <= 1 ? x : fib(x-1) + fib(x-2)
	}
	fib(2)
	`, opts, func(bc *Bytecode) {
		expected := copyBytecode(bc)
		expectPatch(t, bc, expectedPatch{
			bc:            expected,
			noCmpCompFunc: true,
			noCmpConsts:   true,
			numInserts:    2,
			numCalls:      1 + 3,
		})
	})
	expectCompile(t, `
	a := 2
	try {
		throw a
	} catch err {
		for a > 0 {
			a--
		}
	} finally {
		return
	}
	`, opts, func(bc *Bytecode) {
		expected := copyBytecode(bc)
		expectPatch(t, bc, expectedPatch{
			bc:            expected,
			noCmpCompFunc: true,
			noCmpConsts:   true,
			numInserts:    2,
			numCalls:      3,
		})
	})
	expectCompile(t, ``, opts, func(bc *Bytecode) {
		expected := copyBytecode(bc)
		expected.Main.SourceMap = map[int]int{7: 0}
		n, err := patcher.Gosched(bc, 100)
		require.NoError(t, err)
		require.Equal(t, 1, n)
		expected.Main.Instructions = concatInsts(
			makeInst(OpConstant, 0),
			makeInst(OpCall, 0, 0),
			makeInst(OpPop),
			makeInst(OpReturn, 0),
		)
		expectCompiledFunctionsEqual(t, bc.Main, expected.Main)
	})
	expectCompile(t, `
	a := 2
	try {
		throw a
	} catch err {
		for a > 0 {
			a--
		}
	} finally {
		return
	}
	`, opts, func(bc *Bytecode) {
		orig := copyBytecode(bc)
		n, err := patcher.Gosched(bc, 100)
		require.NoError(t, err)
		require.Equal(t, 2, n)
		expected := `
Params:0 Variadic:false Locals:2
Instructions:
0000 CONSTANT        3
0003 CALL            0    0
0006 POP
0007 CONSTANT        0
0010 DEFINELOCAL     0
0012 SETUPTRY        33    69
0021 GETLOCAL        0
0023 THROW           1
0025 NULL
0026 DEFINELOCAL     1
0028 JUMP            69
0033 SETUPCATCH
0034 SETLOCAL        1
0036 GETLOCAL        0
0038 CONSTANT        1
0041 BINARYOP        40
0043 JUMPFALSY       69
0048 GETLOCAL        0
0050 CONSTANT        2
0053 BINARYOP        13
0055 SETLOCAL        0
0057 CONSTANT        3
0060 CALL            0    0
0063 POP
0064 JUMP            36
0069 SETUPFINALLY
0070 RETURN          0
0072 THROW           0
0074 RETURN          0
SourceMap:map[7:8 10:3 12:11 21:25 23:19 25:30 26:11 28:11 33:30 34:30 36:48 38:52 41:48 43:44 48:59 50:60 53:59 55:59 64:44 69:70 70:82 72:11 74:0]
`
		var buf bytes.Buffer
		bc.Main.Fprint(&buf)
		got := buf.String()
		expected = trimLines(expected)
		got = trimLines(got)
		require.Equalf(t, expected, got, "Got:\n%s\n\nOriginal:\n%s", got, orig)
	})
	expectCompile(t, `
	a := 2
	try {
		throw a
	} finally {
		return
	}
	`, opts, func(bc *Bytecode) {
		orig := copyBytecode(bc)
		n, err := patcher.Gosched(bc, 100)
		require.NoError(t, err)
		require.Equal(t, 1, n)
		expected := `
Params:0 Variadic:false Locals:1
Instructions:
0000 CONSTANT        1
0003 CALL            0    0
0006 POP
0007 CONSTANT        0
0010 DEFINELOCAL     0
0012 SETUPTRY        0    25
0021 GETLOCAL        0
0023 THROW           1
0025 SETUPFINALLY
0026 RETURN          0
0028 THROW           0
0030 RETURN          0
SourceMap:map[7:8 10:3 12:11 21:25 23:19 25:30 26:42 28:11 30:0]
`
		var buf bytes.Buffer
		bc.Main.Fprint(&buf)
		got := buf.String()
		expected = trimLines(expected)
		got = trimLines(got)
		require.Equalf(t, expected, got, "Got:\n%s\nOriginal:\n%s", got, orig)
	})
}

func trimLines(s string) string {
	lines := strings.Split(strings.TrimSpace(s), "\n")
	for i := range lines {
		lines[i] = strings.TrimSpace(lines[i])
	}
	return strings.Join(lines, "\n")
}

func concatInsts(insts ...[]byte) []byte {
	var out []byte
	for i := range insts {
		out = append(out, insts[i]...)
	}
	return out
}

func makeInst(op Opcode, args ...int) []byte {
	inst, err := MakeInstruction(make([]byte, 8), op, args...)
	if err != nil {
		panic(err)
	}
	return inst
}

func copyBytecode(bc *Bytecode) *Bytecode {
	return &Bytecode{
		FileSet:    bc.FileSet,
		Constants:  Array(bc.Constants).Copy().(Array),
		Main:       bc.Main.Copy().(*CompiledFunction),
		NumModules: bc.NumModules,
	}
}

type expectedPatch struct {
	bc            *Bytecode
	noExec        bool
	noCmpConsts   bool
	noCmpCompFunc bool
	numInserts    int
	callCount     uint32
	numCalls      uint32
}

func expectCompile(t *testing.T, s string, opts CompilerOptions, fn func(*Bytecode)) {
	t.Helper()
	bc, err := Compile([]byte(s), opts)
	require.NoError(t, err)
	fn(bc)
}

func expectPatch(t *testing.T, actual *Bytecode, expected expectedPatch) {
	t.Helper()
	defer func() {
		if t.Failed() {
			fmt.Printf("Expected: %s\nPatched: %s",
				expected.bc.String(), actual.String())
		}
	}()
	require.NotSame(t, expected.bc, actual, "do not use same object to test")
	cc := uint32(100)
	if expected.callCount > 0 {
		cc = expected.callCount
	}
	numInserts, err := patcher.Gosched(actual, cc)
	require.NoError(t, err, "Gosched error")
	require.Equal(t, expected.numInserts, numInserts,
		"number of inserts not equal")
	obj := actual.Constants[len(actual.Constants)-1]
	require.Equal(t, "<gosched>", obj.String(), "got unexpected String()")
	require.Equal(t, "<gosched>", obj.TypeName(), "got unexpected TypeName()")
	fn := obj.(interface {
		NumCalls() uint32
	})
	if !expected.noExec {
		ret, err := NewVM(actual).Run(nil)
		require.NoError(t, err, "VM error")
		require.Equal(t, Undefined, ret, "tests must return undefined")
		require.Equal(t, expected.numCalls, fn.NumCalls(),
			"number of calls not equal")
	}

	require.Equal(t, expected.bc.FileSet, actual.FileSet, "FileSets not equal")
	require.Equal(t, expected.bc.NumModules, actual.NumModules,
		"number of modules not equal")
	if !expected.noCmpConsts {
		expectConstantsEqual(t,
			expected.noCmpConsts, actual.Constants, expected.bc.Constants)
	}
	if !expected.noCmpCompFunc {
		expectCompiledFunctionsEqual(t, actual.Main, expected.bc.Main)
	}
}

func expectCompiledFunctionsEqual(t *testing.T, actual, expected *CompiledFunction) {
	t.Helper()
	defer func() {
		if t.Failed() {
			var w bytes.Buffer
			expected.Fprint(&w)
			exp := w.String()
			w.Reset()
			actual.Fprint(&w)
			got := w.String()
			fmt.Printf("Expected CompiledFunction:\n%s\nGot:\n%s\n", exp, got)
		}
	}()
	require.Equal(t, expected.Free, actual.Free, "Free not equal")
	require.True(t, string(expected.Instructions) == string(actual.Instructions),
		"Instruction not equal")
	require.Equal(t, expected.NumLocals, actual.NumLocals, "NumLocals not equal")
	require.Equal(t, expected.NumParams, actual.NumParams, "NumParams not equal")
	require.Equal(t, expected.Variadic, actual.Variadic, "Variadic not equal")
	expectSourceMapsEqual(t, actual.SourceMap, expected.SourceMap)
}

func expectConstantsEqual(t *testing.T, noCmp bool, actual, expected []Object) {
	t.Helper()
	require.Equal(t, len(expected)+1, len(actual))
	for i := range expected {
		if f, ok := expected[i].(*CompiledFunction); ok {
			if !noCmp {
				expectCompiledFunctionsEqual(t,
					actual[i].(*CompiledFunction), f)
			}
		} else {
			require.Equalf(t, expected[i], actual[i],
				"constants not equal at %d", i)
		}
	}
}

func expectSourceMapsEqual(t *testing.T, actual, expected map[int]int) {
	require.Equal(t, fmt.Sprint(expected), fmt.Sprint(actual),
		"SourceMap not equal",
	)
}
