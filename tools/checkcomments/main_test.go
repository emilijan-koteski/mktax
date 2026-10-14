package main

import (
	"strings"
	"testing"
)

func TestCheckFile(t *testing.T) {
	cases := []struct {
		name string
		path string
		src  string
		want []string
	}{
		{"go trailing comment", "a.go", "package a\n\nvar x = 1 // the x\n", []string{"a.go:3: trailing comment"}},
		{"go package doc", "a.go", "// Package a does things.\npackage a\n", []string{"a.go:1: comment outside a function header"}},
		{"go field comment", "a.go", "package a\n\ntype T struct {\n\t// the field\n\tF int\n}\n", []string{"a.go:4: comment outside a function header"}},
		{"go two sentence header", "a.go", "package a\n\n// Does the thing. It is complex.\nfunc f() {}\n", nil},
		{"go three sentence header", "a.go", "package a\n\n// One. Two. Three.\nfunc f() {}\n", []string{"a.go:3: function header longer than 2 sentences"}},
		{"go four line header", "a.go", "package a\n\n// a\n// b\n// c\n// d\nfunc f() {}\n", []string{"a.go:3: function header longer than 3 lines"}},
		{"go method header", "a.go", "package a\n\ntype T struct{}\n\n// Explains the method.\nfunc (T) m() {}\n", nil},
		{"go directives", "a.go", "//go:build js && wasm\n\npackage a\n\n//go:embed all:static\nvar f string\n\n//nolint:errcheck\nfunc g() {}\n", nil},
		{"go block comment in body", "a.go", "package a\n\nfunc f() {\n\t/* why */\n}\n", []string{"a.go:4: comment outside a function header"}},
		{"ts stray comment", "a.ts", "const a = 1;\n// stray\nconst b = 2;\n", []string{"a.ts:2: comment outside a function header"}},
		{"ts trailing comment", "a.ts", "const a = 1; // one\n", []string{"a.ts:1: trailing comment"}},
		{"ts url in string", "a.ts", "const u = 'https://example.org';\n", nil},
		{"ts regex with slashes", "a.ts", "const r = /https?:\\/\\//;\n", nil},
		{"ts arrow function header", "a.ts", "// Builds the thing.\nexport const build = (x: number) => x;\n", nil},
		{"ts function header three sentences", "a.ts", "// One. Two. Three.\nfunction f() {}\n", []string{"a.ts:1: function header longer than 2 sentences"}},
		{"ts jsdoc on function", "a.ts", "/**\n * Parses it.\n */\nexport async function parse() {}\n", nil},
		{"ts class method header", "a.ts", "class A {\n  // Runs it.\n  async run(): Promise<void> {}\n}\n", nil},
		{"ts comment above if", "a.ts", "function f(x: number) {\n  // guard\n  if (x) {\n  }\n}\n", []string{"a.ts:2: comment outside a function header"}},
		{"ts directives", "a.ts", "/// <reference types=\"vitest/config\" />\n// eslint-disable-next-line no-console\n// @ts-expect-error missing types\n// @vitest-environment node\nconst a = 1;\n", nil},
		{"ts block comment", "a.ts", "const a = 1;\n/* block */\n", []string{"a.ts:2: comment outside a function header"}},
		{"svelte html comment", "a.svelte", "<script lang=\"ts\">\n  let a = 1;\n</script>\n<!-- layout -->\n<div>{a}</div>\n", []string{"a.svelte:4: comment outside a function header"}},
		{"svelte ignore directive", "a.svelte", "<!-- svelte-ignore a11y_no_static_element_interactions -->\n<div></div>\n", nil},
		{"svelte script trailing comment", "a.svelte", "<script lang=\"ts\">\n  let a = 1; // one\n</script>\n", []string{"a.svelte:2: trailing comment"}},
		{"svelte style comment", "a.svelte", "<style>\n  /* colours */\n  .a { color: red; }\n</style>\n", []string{"a.svelte:2: comment outside a function header"}},
		{"css comment", "a.css", "@import \"tailwindcss\";\n/* tokens */\n", []string{"a.css:2: comment outside a function header"}},
		{"markdown ignored", "a.md", "<!-- anything -->\n", nil},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := render(checkFile(tc.path, []byte(tc.src)))
			if strings.Join(got, "\n") != strings.Join(tc.want, "\n") {
				t.Fatalf("got %q, want %q", got, tc.want)
			}
		})
	}
}

func TestFilePathFromHook(t *testing.T) {
	cases := []struct{ in, want string }{
		{`{"tool_name":"Edit","tool_input":{"file_path":"/x/a.go","old_string":"a","new_string":"b"}}`, "/x/a.go"},
		{`{"tool_name":"Write","tool_input":{"file_path":"/x/b.ts","content":"// c"}}`, "/x/b.ts"},
		{`{"tool_name":"Bash","tool_input":{"command":"ls"}}`, ""},
		{`not json`, ""},
	}
	for _, tc := range cases {
		if got := filePathFromHook([]byte(tc.in)); got != tc.want {
			t.Errorf("%s: got %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestSourceFile(t *testing.T) {
	yes := []string{"a.go", "web/src/x.ts", "x.svelte", "x.js", "x.css"}
	no := []string{"a.md", "a.json", "a.yml", "Makefile", "a_test.go.bak"}
	for _, p := range yes {
		if !sourceFile(p) {
			t.Errorf("%s should be checked", p)
		}
	}
	for _, p := range no {
		if sourceFile(p) {
			t.Errorf("%s should be ignored", p)
		}
	}
}

func TestSkipDir(t *testing.T) {
	for _, d := range []string{"node_modules", "dist", "testdata", ".git", "static", "public", ".superpowers", "bin"} {
		if !skipDir(d) {
			t.Errorf("%s should be skipped", d)
		}
	}
	if skipDir("src") {
		t.Error("src should not be skipped")
	}
}
