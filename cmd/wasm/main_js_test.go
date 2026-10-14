//go:build js && wasm

package main

import (
	"strings"
	"syscall/js"
	"testing"
)

func api(t *testing.T) js.Value {
	t.Helper()
	register()
	v := js.Global().Get("mktax")
	if v.Type() != js.TypeObject {
		t.Fatal("globalThis.mktax is not an object")
	}
	return v
}

func TestRegisterExposesAPI(t *testing.T) {
	v := api(t)
	for _, name := range []string{"parse", "calculate", "version"} {
		if v.Get(name).Type() != js.TypeFunction {
			t.Errorf("mktax.%s is not a function", name)
		}
	}
	if v.Call("version").String() == "" {
		t.Error("empty version")
	}
}

func TestParseViaJS(t *testing.T) {
	out := api(t).Call("parse", `{"files":[{"name":"a.csv","data":"Action,Time (UTC),ISIN,Ticker,Name,No. of shares,Price / share,Currency (Price / share),Total,Currency (Total)\nMarket buy,2026-05-28 13:39:45+00:00,US0000000001,TEST_US_EQ,Test,1,10,USD,9.2,EUR\n"}],"now":"2026-10-14T12:00:00Z"}`).String()
	if !strings.HasPrefix(out, `{"ok":true,`) || !strings.Contains(out, `"trades":1`) {
		t.Fatalf("parse = %s", out)
	}
}

func TestCalculateViaJS(t *testing.T) {
	out := api(t).Call("calculate", `{"files":[],"now":"2026-10-14T12:00:00Z"}`).String()
	if !strings.HasPrefix(out, `{"ok":true,`) || !strings.Contains(out, `"months":[`) {
		t.Fatalf("calculate = %s", out)
	}
}

func TestNonStringArgument(t *testing.T) {
	for _, arg := range []any{5, nil, js.Undefined()} {
		out := api(t).Call("parse", arg).String()
		if !strings.Contains(out, `"ok":false`) || !strings.Contains(out, `"bad_input"`) {
			t.Fatalf("arg %v: %s", arg, out)
		}
	}
	if out := api(t).Call("parse").String(); !strings.Contains(out, `"bad_input"`) {
		t.Fatalf("no args: %s", out)
	}
}
