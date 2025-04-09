//go:build js && wasm
// +build js,wasm

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"sync"
	"syscall/js"
	"time"

	ugofmt "github.com/ozanh/ugo/stdlib/fmt"
	ugojson "github.com/ozanh/ugo/stdlib/json"
	ugostrings "github.com/ozanh/ugo/stdlib/strings"
	ugotime "github.com/ozanh/ugo/stdlib/time"

	"github.com/ozanh/ugo"
	"github.com/ozanh/ugodev/patcher"
)

const maxExecDuration = 60 * time.Second
const playgroundBusy = "ugo playground is busy"

var gStdout = bytes.NewBuffer(nil)
var gMutex sync.Mutex
var gBusy bool
var gRunCancel context.CancelFunc

func init() {
	ugo.PrintWriter = gStdout
}

type Metrics struct {
	start   time.Time
	compile time.Duration
	exec    time.Duration
}

func (m *Metrics) init() {
	m.start = time.Now()
}

func (m *Metrics) initCompile() func() {
	start := time.Now()
	return func() {
		m.compile = time.Since(start)
	}
}

func (m *Metrics) initExec() func() {
	start := time.Now()
	return func() {
		m.exec = time.Since(start)
	}
}

func (m *Metrics) output() map[string]any {
	return map[string]any{
		"elapsed": time.Since(m.start).Round(time.Microsecond).String(),
		"compile": m.compile.Round(time.Microsecond).String(),
		"exec":    m.exec.Round(time.Microsecond).String(),
	}
}

func newResult(err string, value any, metrics map[string]any) map[string]any {
	return map[string]any{
		"stdout":  gStdout.String(),
		"error":   err,
		"value":   value,
		"metrics": metrics,
	}
}

func newErrorResult(err string) map[string]any {
	return map[string]any{
		"stdout": "",
		"error":  err,
	}
}

func makeRunFunc() js.Func {
	opts := ugo.CompilerOptions{
		ModuleMap: ugo.NewModuleMap().
			AddBuiltinModule("time", ugotime.Module).
			AddBuiltinModule("strings", ugostrings.Module).
			AddBuiltinModule("fmt", ugofmt.Module).
			AddBuiltinModule("json", ugojson.Module),
	}

	return js.FuncOf(func(this js.Value, args []js.Value) any {
		if len(args) != 2 {
			return newErrorResult(
				ugo.ErrWrongNumArguments.
					NewError("got =", strconv.Itoa(len(args))).String(),
			)
		}

		gMutex.Lock()
		defer gMutex.Unlock()

		if gBusy {
			return newErrorResult(playgroundBusy)
		}

		metrics := Metrics{}
		metrics.init()

		arg0 := args[0]
		callback := func(v any) { _ = arg0.Call("resultCallback", v) }

		gBusy = true

		ctx, cancel := context.WithCancel(context.Background())
		gRunCancel = cancel

		go func() {
			defer func() {
				gMutex.Lock()
				defer gMutex.Unlock()

				gBusy = false
				gRunCancel = nil
				cancel()
			}()

			defer func() {
				if r := recover(); r != nil {
					result := newResult(fmt.Sprintf("panic: %+v", r), "", metrics.output())
					callback(result)
				}
				if gStdout.Cap() > 64*1024 {
					*gStdout = bytes.Buffer{}
					return
				}
				gStdout.Reset()
			}()

			gStdout.Reset()

			script := args[1].String()
			calcCompTime := metrics.initCompile()
			bc, err := ugo.Compile([]byte(script), opts)
			calcCompTime()
			if err != nil {
				callback(newResult(err.Error(), "", metrics.output()))
				return
			}

			const schedThreshold = 1000
			if _, err = patcher.PatchForGosched(bc, schedThreshold); err != nil {
				callback(newResult(err.Error(), "", metrics.output()))
				return
			}

			var ret ugo.Object
			var waitCh = make(chan struct{})

			vm := ugo.NewVM(bc)
			defer vm.Abort()

			go func() {
				defer close(waitCh)
				defer metrics.initExec()()

				ret, err = vm.Run(nil)
			}()

			tm := time.NewTimer(maxExecDuration)
			defer tm.Stop()

			var isCtxDone bool
			select {
			case <-ctx.Done():
				isCtxDone = true
				vm.Abort()
			case <-tm.C:
				vm.Abort()
			case <-waitCh:
			}

			<-waitCh

			if err != nil {
				if errors.Is(err, ugo.ErrVMAborted) {
					if isCtxDone {
						err = fmt.Errorf("%w %w", err, ctx.Err())
					} else {
						err = fmt.Errorf("%w %s playground max execution time",
							err, maxExecDuration.String())
					}
				}
				e := fmt.Sprintf("%+v", err)
				callback(newResult(e, "", metrics.output()))
				return
			}

			var result map[string]any
			if ret != nil {
				s, err := json.Marshal(objectToAny(ret))
				if err != nil {
					result = newResult(err.Error(), ret.String(), metrics.output())
				} else {
					result = newResult("", string(s), metrics.output())
				}
			} else {
				result = newResult("", "<nil>", metrics.output())
			}
			callback(result)
		}()
		return nil
	})
}

func makeCancelFunc() js.Func {
	return js.FuncOf(func(_ js.Value, _ []js.Value) any {
		gMutex.Lock()
		defer gMutex.Unlock()

		fmt.Println("cancel called")

		if gRunCancel == nil {
			return false
		}

		gRunCancel()
		gRunCancel = nil
		return true
	})
}

func objectToAny(v ugo.Object) any {
	switch vv := v.(type) {
	case ugo.Array:
		arr := make([]any, len(vv))
		for i, mv := range vv {
			arr[i] = objectToAny(mv)
		}
		return arr
	case ugo.Map:
		m := make(map[string]any, len(vv))
		for k, mv := range vv {
			m[k] = objectToAny(mv)
		}
		return m
	case *ugo.CompiledFunction:
		return vv.String()
	case *ugo.Function:
		return vv.String()
	case *ugo.BuiltinFunction:
		return vv.String()
	case *ugo.Error:
		return vv.String()
	case *ugo.RuntimeError:
		return fmt.Sprintf("%+v", vv)
	default:
		if v == ugo.Undefined {
			return nil
		}
		return v
	}
}

func main() {
	fmt.Println("uGo Playground for WebAssembly")

	run := makeRunFunc()
	defer run.Release()

	check := makeCheckFunc()
	defer check.Release()

	cancel := makeCancelFunc()
	defer cancel.Release()

	global := js.Global()

	global.Set("cancelUGO", cancel)
	global.Set("checkUGO", check)
	global.Set("runUGO", run)

	select {}
}
