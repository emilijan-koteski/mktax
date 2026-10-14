package main

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const (
	ruleTrailing  = "trailing comment"
	ruleOutside   = "comment outside a function header"
	ruleSentences = "function header longer than 2 sentences"
	ruleLines     = "function header longer than 3 lines"
	maxSentences  = 2
	maxLines      = 3
)

type violation struct {
	file string
	line int
	rule string
}

type comment struct {
	start, end int
	text       string
	trailing   bool
	single     bool
}

type mode int

const (
	modeScript mode = iota
	modeCSS
	modeHTML
)

var (
	sentenceEnd = regexp.MustCompile(`[.!?](\s|$)`)
	declRes     = []*regexp.Regexp{
		regexp.MustCompile(`^\s*(export\s+)?(default\s+)?(async\s+)?function\b`),
		regexp.MustCompile(`^\s*(export\s+)?(const|let|var)\s+[\w$]+\b.*(=>|=\s*(async\s+)?function\b)`),
		regexp.MustCompile(`^\s*(export\s+)?(default\s+)?(abstract\s+)?class\b`),
		regexp.MustCompile(`^\s*((public|private|protected|static|readonly)\s+)*[\w$]+\s*(:[^=]+)?=\s*(async\s*)?(\([^)]*\)|[\w$]+)\s*=>`),
	}
	methodRe     = regexp.MustCompile(`^\s*((public|private|protected|static|async|override|readonly)\s+)*(get\s+|set\s+)?([A-Za-z_$][\w$]*)\s*(<[^>]*>)?\s*\([^)]*\)\s*(:\s*[^{]+)?\s*(\{.*)?$`)
	keywords     = map[string]bool{"if": true, "for": true, "while": true, "switch": true, "catch": true, "return": true, "function": true, "else": true, "do": true, "try": true, "with": true}
	directives   = []string{"/ <reference", "eslint-disable", "eslint-enable", "@ts-", "@vitest-environment", "prettier-ignore", "svelte-ignore", "@__PURE__"}
	goDirectives = []string{"//go:", "//nolint", "//lint:", "//export ", "// +build", "//sys"}
	skipDirs     = map[string]bool{".git": true, "node_modules": true, "dist": true, "testdata": true, "static": true, "public": true, ".superpowers": true, "bin": true}
	sourceExts   = map[string]bool{".go": true, ".ts": true, ".js": true, ".svelte": true, ".css": true}
)

func main() {
	var files []string
	if len(os.Args) > 1 && os.Args[1] == "-hook" {
		in, _ := io.ReadAll(os.Stdin)
		p := filePathFromHook(in)
		if p == "" || !sourceFile(p) {
			return
		}
		files = []string{p}
	} else {
		for _, a := range os.Args[1:] {
			files = append(files, walk(strings.TrimSuffix(a, "..."))...)
		}
	}
	var all []violation
	for _, f := range files {
		src, err := os.ReadFile(f)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		all = append(all, checkFile(f, src)...)
	}
	for _, l := range render(all) {
		fmt.Fprintln(os.Stderr, l)
	}
	if len(all) > 0 {
		os.Exit(2)
	}
}

func filePathFromHook(b []byte) string {
	var h struct {
		ToolInput struct {
			FilePath string `json:"file_path"`
		} `json:"tool_input"`
	}
	if json.Unmarshal(b, &h) != nil {
		return ""
	}
	return h.ToolInput.FilePath
}

func sourceFile(p string) bool { return sourceExts[filepath.Ext(p)] }

func skipDir(name string) bool { return skipDirs[name] }

func walk(root string) []string {
	var out []string
	_ = filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			if p != root && skipDir(d.Name()) {
				return filepath.SkipDir
			}
			return nil
		}
		if sourceFile(p) {
			out = append(out, p)
		}
		return nil
	})
	return out
}

func render(vs []violation) []string {
	out := make([]string, 0, len(vs))
	for _, v := range vs {
		out = append(out, fmt.Sprintf("%s:%d: %s", v.file, v.line, v.rule))
	}
	return out
}

func checkFile(path string, src []byte) []violation {
	switch filepath.Ext(path) {
	case ".go":
		return checkGo(path, src)
	case ".ts", ".js":
		return checkScript(path, string(src), modeScript)
	case ".css":
		return checkScript(path, string(src), modeCSS)
	case ".svelte":
		return checkScript(path, string(src), modeHTML)
	}
	return nil
}

func checkGo(path string, src []byte) []violation {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, path, src, parser.ParseComments|parser.SkipObjectResolution)
	if err != nil {
		return []violation{{path, 1, "parse error: " + err.Error()}}
	}
	docs := map[*ast.CommentGroup]bool{}
	for _, d := range f.Decls {
		if fd, ok := d.(*ast.FuncDecl); ok && fd.Doc != nil {
			docs[fd.Doc] = true
		}
	}
	lines := strings.Split(string(src), "\n")
	var out []violation
	for _, cg := range f.Comments {
		if goDirective(cg) {
			continue
		}
		pos := fset.Position(cg.Pos())
		switch {
		case docs[cg]:
			if r := headerRule(cg.Text(), fset.Position(cg.End()).Line-pos.Line+1); r != "" {
				out = append(out, violation{path, pos.Line, r})
			}
		case strings.TrimSpace(lines[pos.Line-1][:pos.Column-1]) != "":
			out = append(out, violation{path, pos.Line, ruleTrailing})
		default:
			out = append(out, violation{path, pos.Line, ruleOutside})
		}
	}
	return out
}

func goDirective(cg *ast.CommentGroup) bool {
	for _, c := range cg.List {
		ok := false
		for _, d := range goDirectives {
			if strings.HasPrefix(c.Text, d) {
				ok = true
				break
			}
		}
		if !ok {
			return false
		}
	}
	return true
}

func headerRule(text string, lines int) string {
	if sentences(text) > maxSentences {
		return ruleSentences
	}
	if lines > maxLines {
		return ruleLines
	}
	return ""
}

func sentences(text string) int {
	n := len(sentenceEnd.FindAllStringIndex(strings.TrimSpace(text), -1))
	if n == 0 {
		return 1
	}
	return n
}

func checkScript(path, src string, m mode) []violation {
	lines := strings.Split(src, "\n")
	var out []violation
	for _, c := range merge(scan(src, m)) {
		if directive(c.text) {
			continue
		}
		switch {
		case c.trailing:
			out = append(out, violation{path, c.start, ruleTrailing})
		case declStart(nextCode(lines, c.end)):
			if r := headerRule(cleanText(c.text), c.end-c.start+1); r != "" {
				out = append(out, violation{path, c.start, r})
			}
		default:
			out = append(out, violation{path, c.start, ruleOutside})
		}
	}
	return out
}

func directive(text string) bool {
	for _, l := range strings.Split(text, "\n") {
		l = strings.TrimSpace(strings.TrimLeft(strings.TrimSpace(l), "*"))
		if l == "" {
			continue
		}
		ok := false
		for _, d := range directives {
			if strings.HasPrefix(l, d) {
				ok = true
				break
			}
		}
		if !ok {
			return false
		}
	}
	return true
}

func cleanText(text string) string {
	var b strings.Builder
	for _, l := range strings.Split(text, "\n") {
		b.WriteString(strings.TrimSpace(strings.TrimLeft(strings.TrimSpace(l), "*")))
		b.WriteByte(' ')
	}
	return b.String()
}

func nextCode(lines []string, after int) string {
	for i := after; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) != "" {
			return lines[i]
		}
	}
	return ""
}

func declStart(line string) bool {
	for _, re := range declRes {
		if re.MatchString(line) {
			return true
		}
	}
	if m := methodRe.FindStringSubmatch(line); m != nil {
		return !keywords[m[4]]
	}
	return false
}

func merge(cs []comment) []comment {
	var out []comment
	for _, c := range cs {
		if n := len(out); n > 0 && out[n-1].single && c.single && !c.trailing && !out[n-1].trailing && out[n-1].end+1 == c.start {
			out[n-1].end = c.end
			out[n-1].text += "\n" + c.text
			continue
		}
		out = append(out, c)
	}
	return out
}

// Walks the source once, switching between markup, script and style regions for Svelte files. It records every comment with its line span and whether code preceded it on the same line.
func scan(src string, m mode) []comment {
	var out []comment
	cur := m
	line, i, n := 1, 0, len(src)
	code := false
	var prev byte
	for i < n {
		c := src[i]
		if c == '\n' {
			line++
			code = false
			i++
			continue
		}
		if cur == modeHTML {
			switch {
			case strings.HasPrefix(src[i:], "<!--"):
				j := strings.Index(src[i+4:], "-->")
				if j < 0 {
					j = n - i - 4
				}
				text := src[i+4 : i+4+j]
				end := line + strings.Count(text, "\n")
				out = append(out, comment{line, end, text, code, false})
				line = end
				i += 4 + j + 3
			case hasPrefixFold(src[i:], "<script"):
				i = afterTag(src, i)
				cur = modeScript
				code = false
			case hasPrefixFold(src[i:], "<style"):
				i = afterTag(src, i)
				cur = modeCSS
				code = false
			default:
				if c != ' ' && c != '\t' && c != '\r' {
					code = true
				}
				i++
			}
			continue
		}
		if m == modeHTML && cur == modeScript && hasPrefixFold(src[i:], "</script>") {
			cur = modeHTML
			i += len("</script>")
			continue
		}
		if m == modeHTML && cur == modeCSS && hasPrefixFold(src[i:], "</style>") {
			cur = modeHTML
			i += len("</style>")
			continue
		}
		switch {
		case c == '/' && i+1 < n && src[i+1] == '*':
			j := strings.Index(src[i+2:], "*/")
			if j < 0 {
				j = n - i - 2
			}
			text := src[i+2 : i+2+j]
			end := line + strings.Count(text, "\n")
			out = append(out, comment{line, end, text, code, false})
			line = end
			i += 2 + j + 2
		case cur == modeScript && c == '/' && i+1 < n && src[i+1] == '/':
			j := strings.IndexByte(src[i:], '\n')
			if j < 0 {
				j = n - i
			}
			out = append(out, comment{line, line, src[i+2 : i+j], code, true})
			i += j
		case c == '\'' || c == '"' || (cur == modeScript && c == '`'):
			i = skipString(src, i, &line)
			code, prev = true, c
		case cur == modeScript && c == '/' && regexAllowed(prev):
			i = skipRegex(src, i)
			code, prev = true, '/'
		default:
			if c != ' ' && c != '\t' && c != '\r' {
				code, prev = true, c
			}
			i++
		}
	}
	return out
}

func hasPrefixFold(s, prefix string) bool {
	return len(s) >= len(prefix) && strings.EqualFold(s[:len(prefix)], prefix)
}

func afterTag(src string, i int) int {
	j := strings.IndexByte(src[i:], '>')
	if j < 0 {
		return len(src)
	}
	return i + j + 1
}

func skipString(src string, i int, line *int) int {
	q := src[i]
	for i++; i < len(src); i++ {
		switch src[i] {
		case '\\':
			i++
		case '\n':
			if q != '`' {
				return i
			}
			*line++
		case q:
			return i + 1
		}
	}
	return len(src)
}

func regexAllowed(prev byte) bool {
	return prev == 0 || strings.IndexByte("(,=:[!&|?{};+-*%<>~^", prev) >= 0
}

func skipRegex(src string, i int) int {
	inClass := false
	for i++; i < len(src); i++ {
		switch src[i] {
		case '\\':
			i++
		case '\n':
			return i
		case '[':
			inClass = true
		case ']':
			inClass = false
		case '/':
			if !inClass {
				return i + 1
			}
		}
	}
	return len(src)
}
