package tsserver

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"testing"
	"time"
)

type fileInfo struct{}

func (fileInfo) Name() string { return "" }

func (fileInfo) Size() int64 { return 0 }

func (fileInfo) ModTime() time.Time { return time.Now() }

func (fileInfo) Mode() fs.FileMode { return 0o755 }

func (fileInfo) IsDir() bool { return false }

func (fileInfo) Sys() any { return nil }

type pseudoFS string

func (p pseudoFS) Open(name string) (fs.File, error) {
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
	for n, test := range [...]struct {
		Input, Output string
	}{
		{
			"",
			"",
		},
		{
			"const a: number = 1;\n\nconsole.log(a);",
			"const a/*: number*/ = 1;\n\nconsole.log(a);",
		},
		{
			"const a: = 2;",
			"console.log(\"ModuleItem: error at position 1 (1:1):\\nStatementListItem: error at position 1 (1:1):\\nDeclaration: error at position 1 (1:1):\\nLexicalDeclaration: error at position 8 (1:8):\\ninvalid lexical declaration\")",
		},
	} {
		f, err := WrapFSWithErrorHandler(pseudoFS(test.Input), errFn).Open("a.js")
		if err != nil {
			t.Errorf("test %d: unexpected error while getting file: %s", n+1, err)
		} else if b, _ := io.ReadAll(f); string(b) != test.Output {
			t.Errorf("test %d: expecting output %s, got %s", n+1, test.Output, b)
		}
	}
}
