package tsserver

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"path"
	"testing"
	"text/template"
	"time"
)

type fileInfo struct{}

func (fileInfo) Name() string { return "" }

func (fileInfo) Size() int64 { return 0 }

func (fileInfo) ModTime() time.Time { return time.Now() }

func (fileInfo) Mode() fs.FileMode { return 0o755 }

func (fileInfo) IsDir() bool { return false }

func (fileInfo) Sys() any { return nil }

type pseudoFS [4]string

func (pf pseudoFS) Open(name string) (fs.File, error) {
	p := pf[0]

	switch path.Ext(name) {
	case tsExt:
		p = pf[1]
	case jsxExt:
		p = pf[2]
	case tsxExt:
		p = pf[3]
	}

	if p == "" {
		return nil, fs.ErrNotExist
	}

	return &file{
		Reader:   bytes.NewReader([]byte(p)),
		name:     name,
		FileInfo: fileInfo{},
	}, nil
}

func errFn(w io.Writer, err error) {
	fmt.Fprintf(w, "console.log(%q)", err)
}

func TestWrap(t *testing.T) {
	jsx := template.Must(template.New("").Parse("import {createElement} from '@dom'; createElement('TAG_NAME', PARAMS, CHILDREN)"))
	for n, test := range [...]struct {
		Input  pseudoFS
		Output string
		HasJSX bool
	}{
		{ // 1
			pseudoFS{"a", "", "", ""},
			"a",
			false,
		},
		{ // 2
			pseudoFS{"", "const a: number = 1;\n\nconsole.log(a);", "", ""},
			"const a/*: number*/ = 1;\n\nconsole.log(a);",
			false,
		},
		{ // 3
			pseudoFS{"", "const a: = 2;", "", ""},
			"console.log(\"ModuleItem: error at position 1 (1:1):\\nStatementListItem: error at position 1 (1:1):\\nDeclaration: error at position 1 (1:1):\\nLexicalDeclaration: error at position 8 (1:8):\\ninvalid lexical declaration\")",
			false,
		},
		{ // 4
			pseudoFS{"", "const a = 123", "const a = <div />", ""},
			"const a = 123",
			false,
		},
		{ // 5
			pseudoFS{"", "const a = 123", "const a = <div />", ""},
			"import{createElement}from\"@dom\"\nconst a = (createElement(\"div\", {}, []))",
			true,
		},
		{ // 6
			pseudoFS{"", "const a = 123", "const a = <div />", "const b = <div />"},
			"import{createElement}from\"@dom\"\nconst b = (createElement(\"div\", {}, []))",
			true,
		},
		{ // 7
			pseudoFS{"", "const a = 123", "const a = <div />", "const b = <div />"},
			"const a = 123",
			false,
		},
	} {
		opts := make([]Option, 1, 2)

		opts[0] = ErrFn(errFn)

		if test.HasJSX {
			opts = append(opts, JSX(jsx))
		}

		f, err := WrapFS(test.Input, opts...).Open("a.js")
		if err != nil {
			t.Errorf("test %d: unexpected error while getting file: %s", n+1, err)
		} else if b, _ := io.ReadAll(f); string(b) != test.Output {
			t.Errorf("test %d: expecting output %s, got %s", n+1, test.Output, b)
		}
	}
}
