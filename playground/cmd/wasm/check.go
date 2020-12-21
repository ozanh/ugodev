// +build js,wasm

package main

import (
	"fmt"
	"strconv"
	"syscall/js"

	"github.com/ozanh/ugo"
	"github.com/ozanh/ugo/parser"
	ugofmt "github.com/ozanh/ugo/stdlib/fmt"
	ugostrings "github.com/ozanh/ugo/stdlib/strings"
	ugotime "github.com/ozanh/ugo/stdlib/time"
)

// linesErrors returns line numbers and assoc. error messages thrown by parser,
// optimizer, compiler or VM.
func linesErrors(err error) map[string]interface{} {
	m := linesFromError(err)
	if len(m) > 0 {
		um := uniqueLinesErrorsString(m)
		out := make(map[string]interface{}, len(um))
		for k, v := range um {
			l := make([]interface{}, len(v))
			for i, s := range v {
				l[i] = s
			}
			out[strconv.Itoa(k)] = l
		}
		return out
	}
	return nil
}

func linesFromError(err error) map[int][]error {

	switch v := err.(type) {
	case parser.ErrorList:
		out := make(map[int][]error)
		for i := range v {
			out[v[i].Pos.Line] = append(out[v[i].Pos.Line], v[i])
		}
		return out
	case *parser.Error:
		return map[int][]error{
			v.Pos.Line: {v},
		}
	case *ugo.CompilerError:
		if v.FileSet == nil || v.Node == nil {
			return nil
		}
		return map[int][]error{
			v.FileSet.Position(v.Node.Pos()).Line: {v},
		}
	case *ugo.OptimizerError:
		return map[int][]error{
			v.FilePos.Line: {v},
		}
	case interface{ Errors() []error }: // optimizer multipleErr implements this
		errs := v.Errors()
		out := make(map[int][]error)
		for i := range errs {
			for k, vv := range linesFromError(errs[i]) {
				out[k] = append(out[k], vv...)
			}
		}
		return out
	case *ugo.RuntimeError:
		out := make(map[int][]error)
		for _, tr := range v.StackTrace() {
			out[tr.Line] = append(out[tr.Line], v)
		}
		return out
	case interface{ Unwrap() error }:
		return linesFromError(v.Unwrap())
	}
	return nil
}

func uniqueLinesErrorsString(errmap map[int][]error) map[int][]string {
	out := make(map[int][]string, len(errmap))
	for k, v := range errmap {
		out[k] = uniqueErrorStrings(v)
	}
	return out
}

func uniqueErrorStrings(errs []error) []string {
	m := make(map[string]struct{}, len(errs))
	for i := range errs {
		m[errs[i].Error()] = struct{}{}
	}
	out := make([]string, len(m))
	i := 0
	for k := range m {
		out[i] = k
		i++
	}
	return out
}

func newCheckResult(
	warning string,
	linesErrs map[string]interface{},
) map[string]interface{} {
	return map[string]interface{}{
		"warning": warning,
		"lines":   linesErrs,
	}
}

// checkWrapper returns a js function to report given script whether has parse
// and compile errors. Result of check is sent via a callback in this format
// {"warning": <string>, "lines": {<string>: [<string>]}}
func checkWrapper() js.Func {
	opts := ugo.DefaultCompilerOptions
	opts.ModuleMap = ugo.NewModuleMap().
		AddBuiltinModule("time", ugotime.Module).
		AddBuiltinModule("strings", ugostrings.Module).
		AddBuiltinModule("fmt", ugofmt.Module)

	return js.FuncOf(func(this js.Value, args []js.Value) (value interface{}) {
		if len(args) != 2 {
			return newCheckResult(ugo.ErrWrongNumArguments.
				NewError("got =", strconv.Itoa(len(args))).String(), nil)
		}

		callback := func(v interface{}) {
			_ = args[0].Call("checkCallback", v)
		}
		script := args[1].String()
		if script == "" {
			callback(newCheckResult("empty script", nil))
			return nil
		}
		go func() {
			var warning string
			var result map[string]interface{}
			defer func() {
				if r := recover(); r != nil {
					warning = fmt.Sprintf("%+v", r)
				}
				callback(newCheckResult(warning, result))
			}()
			_, err := ugo.Compile([]byte(script), opts)
			if err == nil {
				return
			}
			result = linesErrors(err)
			if result == nil {
				warning = err.Error()
			}
			return
		}()
		return nil
	})
}
