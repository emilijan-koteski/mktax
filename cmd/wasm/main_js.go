//go:build js && wasm

package main

import (
	"syscall/js"

	"github.com/emilijan-koteski/mktax/internal/bridge"
)

const badArgument = `{"ok":false,"error":{"code":"bad_input","message":"expected one JSON string argument"}}`

func main() {
	register()
	select {}
}

func register() {
	api := js.Global().Get("Object").New()
	api.Set("parse", js.FuncOf(func(_ js.Value, args []js.Value) any { return call(args, bridge.ParseJSON) }))
	api.Set("calculate", js.FuncOf(func(_ js.Value, args []js.Value) any { return call(args, bridge.CalculateJSON) }))
	api.Set("version", js.FuncOf(func(js.Value, []js.Value) any { return bridge.Version() }))
	js.Global().Set("mktax", api)
	if ready := js.Global().Get("onMktaxReady"); ready.Type() == js.TypeFunction {
		ready.Invoke()
	}
}

func call(args []js.Value, fn func([]byte) []byte) any {
	if len(args) != 1 || args[0].Type() != js.TypeString {
		return badArgument
	}
	return string(fn([]byte(args[0].String())))
}
