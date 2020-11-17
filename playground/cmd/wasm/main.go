// +build js,wasm

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"syscall/js"
	"time"

	"github.com/ozanh/ugo"
	ugostrings "github.com/ozanh/ugo/stdlib/strings"
	ugotime "github.com/ozanh/ugo/stdlib/time"
)

var stdout = bytes.NewBuffer(nil)

func init() {
	ugo.PrintWriter = stdout
}

type metrics struct {
	start   time.Time
	elapsed time.Duration
	compile time.Duration
	exec    time.Duration
}

func (m *metrics) init() {
	m.start = time.Now()
}

func (m *metrics) initCompile() func() {
	start := time.Now()
	return func() {
		m.compile = time.Since(start)
	}
}

func (m *metrics) initExec() func() {
	start := time.Now()
	return func() {
		m.exec = time.Since(start)
	}
}

func (m *metrics) output() map[string]interface{} {
	return map[string]interface{}{
		"elapsed": time.Since(m.start).Round(time.Microsecond).String(),
		"compile": m.compile.Round(time.Microsecond).String(),
		"exec":    m.exec.Round(time.Microsecond).String(),
	}
}

func newResult(
	err string,
	value interface{},
	metrics map[string]interface{},
) map[string]interface{} {
	return map[string]interface{}{
		"stdout":  stdout.String(),
		"error":   err,
		"value":   value,
		"metrics": metrics,
	}
}

func wrapper() js.Func {
	return js.FuncOf(func(this js.Value, args []js.Value) (value interface{}) {
		mt := metrics{}
		mt.init()
		if len(args) != 2 {
			return newResult(ugo.ErrWrongNumArguments.
				NewError("got =", strconv.Itoa(len(args))).String(), "", mt.output())
		}

		callback := func(v interface{}) {
			_ = args[0].Call("resultCallback", v)
		}

		go func() {
			stdout.Reset()
			defer func() {
				if r := recover(); r != nil {
					callback(newResult(fmt.Sprintf("panic: %v", r), "", mt.output()))
				}
				stdout.Reset()
			}()

			script := args[1].String()
			mm := ugo.NewModuleMap()
			mm.AddBuiltinModule("time", ugotime.Module)
			mm.AddBuiltinModule("strings", ugostrings.Module)
			opts := ugo.DefaultCompilerOptions
			opts.ModuleMap = mm
			f := mt.initCompile()
			bc, err := ugo.Compile([]byte(script), opts)
			f()
			if err != nil {
				callback(newResult(err.Error(), "", mt.output()))
				return
			}
			var ret ugo.Object
			var waitCh = make(chan struct{})
			vm := ugo.NewVM(bc)
			defer vm.Abort()
			go func() {
				defer close(waitCh)
				defer mt.initExec()()
				ret, err = vm.Run(nil)
			}()
			tm := time.NewTimer(10 * time.Second)
			defer tm.Stop()
			select {
			case <-tm.C:
				vm.Abort()
			case <-waitCh:
			}
			<-waitCh
			if err != nil {
				e := fmt.Sprintf("%+v", err)
				callback(newResult(e, "", mt.output()))
				return
			}
			if ret != nil {
				s, err := ugoToJSON(ret)
				if err != nil {
					callback(newResult(err.Error(), ret.String(), mt.output()))
					return
				}
				callback(newResult("", string(s), mt.output()))
				return
			}
			callback(newResult("", "<nil>", mt.output()))
		}()
		return nil
	})
}

func ugoToJSON(v ugo.Object) ([]byte, error) {
	return json.Marshal(conv(v))
}

func conv(v ugo.Object) interface{} {
	switch vv := v.(type) {
	case ugo.Array:
		arr := make([]interface{}, len(vv))
		for i, mv := range vv {
			arr[i] = conv(mv)
		}
		return arr
	case ugo.Map:
		m := make(map[string]interface{}, len(vv))
		for k, mv := range vv {
			m[k] = conv(mv)
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
	w := wrapper()
	defer w.Release()
	js.Global().Set("runUGO", w)
	<-make(chan bool)
	fmt.Println("uGo Playground Stopped")
}
