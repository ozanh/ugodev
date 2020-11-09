package main

import (
	"io/ioutil"
	"syscall/js"
	"testing"
	"time"
)

func setup(t *testing.T) <-chan []js.Value {
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

	w := wrapper()
	t.Cleanup(w.Release)
	global.Set("runUGO", w)
	t.Cleanup(func() { global.Delete("runUGO") })
	return cbArgs
}

func TestUGORun(t *testing.T) {
	global := js.Global()
	cbArgs := setup(t)
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

func TestSample(t *testing.T) {
	global := js.Global()
	code, err := ioutil.ReadFile("testdata/sample.ugo")
	if err != nil {
		t.Fatal(err)
	}
	cbArgs := setup(t)
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
