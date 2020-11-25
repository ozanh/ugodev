// +build js,wasm

package main

import (
	"io/ioutil"
	"strings"
	"syscall/js"
	"testing"
	"time"
)

func setupRun(t *testing.T) <-chan []js.Value {
	t.Helper()
	global := js.Global()
	v := global.Get("_resultCallback")
	if v.Type() != js.TypeUndefined {
		t.Fatal("_resultCallback already set")
	}
	cbArgs := make(chan []js.Value, 1)
	cb := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		cbArgs <- args
		return nil
	})
	t.Cleanup(cb.Release)
	global.Set("_resultCallback", cb)
	t.Cleanup(func() { global.Delete("_resultCallback") })

	global.Call("eval", `var obj = { resultCallback: _resultCallback };`)

	w := runWrapper()
	t.Cleanup(w.Release)
	global.Set("runUGO", w)
	t.Cleanup(func() { global.Delete("runUGO") })
	return cbArgs
}

func setupCheck(t *testing.T) <-chan []js.Value {
	t.Helper()
	global := js.Global()
	v := global.Get("_checkCallback")
	if v.Type() != js.TypeUndefined {
		t.Fatal("_checkCallback already set")
	}
	cbArgs := make(chan []js.Value, 1)
	cb := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		cbArgs <- args
		return nil
	})
	t.Cleanup(cb.Release)
	global.Set("_checkCallback", cb)
	t.Cleanup(func() { global.Delete("_checkCallback") })

	global.Call("eval", `var obj = { checkCallback: _checkCallback };`)

	w := checkWrapper()
	t.Cleanup(w.Release)
	global.Set("checkUGO", w)
	t.Cleanup(func() { global.Delete("checkUGO") })
	return cbArgs
}

func TestUGORun(t *testing.T) {
	global := js.Global()
	cbArgs := setupRun(t)
	v := global.Get("runUGO").Invoke(global.Get("obj"), `x:=123;println(123);return x;`)
	if v.Type() != js.TypeNull {
		t.Fatalf("runUGO() expected: %v, got: %v", js.Null(), v)
	}
	select {
	case args := <-cbArgs:
		if len(args) != 1 {
			t.Fatalf("expected args len: 1, got: %d", len(args))
		}
		if typ := args[0].Type(); typ != js.TypeObject {
			t.Fatalf("expected arg type: %s, got: %s", js.TypeObject, typ)
		}
		if s := args[0].Get("error").String(); s != "" {
			t.Fatalf("expected no error but got: %s", s)
		}
		if s := args[0].Get("stdout").String(); s != "123\n" {
			t.Fatalf("expected stdout: \"123\\n\", got: %q", s)
		}
		if s := args[0].Get("value").String(); s != "123" {
			t.Fatalf("expected value: \"123\", got: %q", s)
		}
		metrics := args[0].Get("metrics")
		if typ := metrics.Type(); typ != js.TypeObject {
			t.Fatalf("expected metrics type: %s, got: %q", js.TypeObject, typ)
		}
		if s := metrics.Get("elapsed").String(); s == "" {
			t.Fatalf("expected metrics.elapsed not empty, got: %q", s)
		}
		if s := metrics.Get("compile").String(); s == "" {
			t.Fatalf("expected metrics.compile not empty, got: %q", s)
		}
		if s := metrics.Get("exec").String(); s == "" {
			t.Fatalf("expected metrics.exec not empty, got: %q", s)
		}
	case <-time.After(time.Second):
		t.Fatal("callback result timeout")
	}
}

func TestUGOCheck(t *testing.T) {
	global := js.Global()
	cbArgs := setupCheck(t)

	// no args return an object with warning
	v := global.Get("checkUGO").Invoke()
	if v.Type() != js.TypeObject {
		t.Fatalf("checkUGO() expected: %v, got: %v", js.TypeObject, v)
	}
	warning := v.Get("warning").String()
	if !strings.HasPrefix(warning, "WrongNumberOfArgumentsError:") {
		t.Fatalf("wrong warning message: %s", warning)
	}

	v = global.Get("checkUGO").Invoke(global.Get("obj"), `x:=123;println(123);return x;`)
	if v.Type() != js.TypeNull {
		t.Fatalf("checkUGO() expected: %v, got: %v", js.Null(), v)
	}
	select {
	case args := <-cbArgs:
		if len(args) != 1 {
			t.Fatalf("expected args len: 1, got: %d", len(args))
		}
		if typ := args[0].Type(); typ != js.TypeObject {
			t.Fatalf("expected arg type: %s, got: %s", js.TypeObject, typ)
		}
		if s := args[0].Get("warning").String(); s != "" {
			t.Fatalf("expected empty warning but got: %s", s)
		}
		if l := args[0].Get("lines"); l.Type() != js.TypeObject {
			t.Fatalf("expected lines type %s, got: %s", js.TypeObject, l.Type())
		}
	case <-time.After(time.Second):
		t.Fatal("callback result timeout")
	}

	// parser error
	v = global.Get("checkUGO").Invoke(global.Get("obj"), "var a,\ntry {}")
	if v.Type() != js.TypeNull {
		t.Fatalf("checkUGO() expected: %v, got: %v", js.Null(), v)
	}
	select {
	case args := <-cbArgs:
		if len(args) != 1 {
			t.Fatalf("expected args len: 1, got: %d", len(args))
		}
		if typ := args[0].Type(); typ != js.TypeObject {
			t.Fatalf("expected arg type: %s, got: %s", js.TypeObject, typ)
		}
		if s := args[0].Get("warning").String(); s != "" {
			t.Fatalf("expected empty warning but got: %s", s)
		}

		arr := args[0].Get("lines").Get("1") // line 1
		if arr.Type() != js.TypeObject {
			t.Fatalf("expected lines[\"1\"] type: %s, got: %s",
				js.TypeObject, arr.Type())
		}
		if arr.Length() != 1 {
			t.Fatalf("expected lines[\"1\"] length: 1, got: %d", arr.Length())
		}
		errMsg := arr.Index(0).String()
		if errMsg != "Parse Error: expected ';', found ','\n\tat (main):1:6" {
			t.Fatal(errMsg)
		}
		arr = args[0].Get("lines").Get("2") // line 2
		if arr.Type() != js.TypeObject {
			t.Fatalf("expected lines[\"2\"] type: %s, got: %s",
				js.TypeObject, arr.Type())
		}
		if arr.Length() != 1 {
			t.Fatalf("expected lines[\"2\"] length: 1, got: %d", arr.Length())
		}
		errMsg = arr.Index(0).String()
		if errMsg != "Parse Error: expected 'finally', found newline\n\tat (main):2:7" {
			t.Fatalf("%q", errMsg)
		}
	case <-time.After(time.Second):
		t.Fatal("callback result timeout")
	}

	// compiler error
	v = global.Get("checkUGO").Invoke(global.Get("obj"), "x=123")
	if v.Type() != js.TypeNull {
		t.Fatalf("checkUGO() expected: %v, got: %v", js.Null(), v)
	}
	select {
	case args := <-cbArgs:
		if len(args) != 1 {
			t.Fatalf("expected args len: 1, got: %d", len(args))
		}
		if typ := args[0].Type(); typ != js.TypeObject {
			t.Fatalf("expected arg type: %s, got: %s", js.TypeObject, typ)
		}
		if s := args[0].Get("warning").String(); s != "" {
			t.Fatalf("expected empty warning but got: %s", s)
		}

		arr := args[0].Get("lines").Get("1") // line 1
		if arr.Type() != js.TypeObject {
			t.Fatalf("expected lines[\"1\"] type: %s, got: %s",
				js.TypeObject, arr.Type())
		}
		if arr.Length() != 1 {
			t.Fatalf("expected lines[\"1\"] length: 1, got: %d", arr.Length())
		}
		errMsg := arr.Index(0).String()
		if errMsg != "Compile Error: unresolved reference \"x\"\n\tat (main):1:1" {
			t.Fatalf("%q", errMsg)
		}
	case <-time.After(time.Second):
		t.Fatal("callback result timeout")
	}

	// optimizer error
	v = global.Get("checkUGO").Invoke(global.Get("obj"), "1/0\n1/0")
	if v.Type() != js.TypeNull {
		t.Fatalf("checkUGO() expected: %v, got: %v", js.Null(), v)
	}
	select {
	case args := <-cbArgs:
		if len(args) != 1 {
			t.Fatalf("expected args len: 1, got: %d", len(args))
		}
		if typ := args[0].Type(); typ != js.TypeObject {
			t.Fatalf("expected arg type: %s, got: %s", js.TypeObject, typ)
		}
		if s := args[0].Get("warning").String(); s != "" {
			t.Fatalf("expected empty warning but got: %s", s)
		}

		arr := args[0].Get("lines").Get("1") // line 1
		if arr.Type() != js.TypeObject {
			t.Fatalf("expected lines[\"1\"] type: %s, got: %s",
				js.TypeObject, arr.Type())
		}
		if arr.Length() != 1 {
			t.Fatalf("expected lines[\"1\"] length: 1, got: %d", arr.Length())
		}
		errMsg := arr.Index(0).String()
		if errMsg != "Optimizer Error: ZeroDivisionError: \n\tat (main):1:1" {
			t.Fatalf("%q", errMsg)
		}

		arr = args[0].Get("lines").Get("2") // line 2
		if arr.Type() != js.TypeObject {
			t.Fatalf("expected lines[\"2\"] type: %s, got: %s",
				js.TypeObject, arr.Type())
		}
		if arr.Length() != 1 {
			t.Fatalf("expected lines[\"2\"] length: 1, got: %d", arr.Length())
		}
		errMsg = arr.Index(0).String()
		if errMsg != "Optimizer Error: ZeroDivisionError: \n\tat (main):2:1" {
			t.Fatalf("%q", errMsg)
		}
	case <-time.After(time.Second):
		t.Fatal("callback result timeout")
	}
}

func TestSample(t *testing.T) {
	global := js.Global()
	code, err := ioutil.ReadFile("testdata/sample.ugo")
	if err != nil {
		t.Fatal(err)
	}
	cbArgs := setupRun(t)
	v := global.Get("runUGO").Invoke(global.Get("obj"), string(code))
	if v.Type() != js.TypeNull {
		t.Fatalf("runUGO() expected: %v, got: %v", js.Null(), v)
	}
	select {
	case args := <-cbArgs:
		if len(args) != 1 {
			t.Fatalf("expected args len: 1, got: %d", len(args))
		}
		if typ := args[0].Type(); typ != js.TypeObject {
			t.Fatalf("expected arg type: %s, got: %s", js.TypeObject, typ)
		}
		if s := args[0].Get("error").String(); s != "" {
			t.Fatalf("expected no error but got: %s", s)
		}
	case <-time.After(time.Second):
		t.Fatal("callback result timeout")
	}
}
