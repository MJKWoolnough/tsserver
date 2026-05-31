// Package tsserver implements a simple wrapper around fs.FS to intercept calls
// to open JavaScript files, and instead open their Typescript equivalents and
// generate the JavaScript.
package tsserver // import "vimagination.zapto.org/tsserver"

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"strings"
	"text/template"

	"vimagination.zapto.org/javascript"
	"vimagination.zapto.org/javascript/jsx"
	"vimagination.zapto.org/parser"
)

const (
	jsExt  = ".js"
	tsExt  = ".ts"
	jsxExt = ".jsx"
	tsxExt = ".tsx"
)

type wrapped struct {
	fs.FS
	errFn func(w io.Writer, err error)
	jsx   *template.Template
}

// WrapFS takes a fs.FS and intercepts any calls to open .js files, and instead
// generates a file from a similarly named .ts file, if one exists.
//
// If a .ts file does not exists, fails to be converted to JavaScript, or if the
// file being opened is not a .js file then the file open will not be
// intercepted.
//
// Deprecated Typescript constructs, such as namespaces and enums are not
// transpiled and will result in a parsing error.
//
// Due to limitation of transpiling a single file at a time, imported types
// must be annotated as type imports, either using the `import type` statement,
// or by prefixing the import identified with the `type` keyword. Anything else
// may result in invalid import errors.
func WrapFS(f fs.FS, opts ...Option) fs.FS {
	wr := &wrapped{FS: f}

	for _, opt := range opts {
		opt(wr)
	}

	return wr
}

// WrapFSWithErrorHandler acts like WrapFS but allows for custom error handling
// via the supplied callback function.
//
// Any data written to the writer will supplied to the client.
func WrapFSWithErrorHandler(f fs.FS, errFn func(w io.Writer, err error)) fs.FS {
	return &wrapped{FS: f, errFn: errFn}
}

func (w *wrapped) Open(name string) (fs.File, error) {
	if before, ok := strings.CutSuffix(name, jsExt); ok {
		var (
			buf  *bytes.Buffer
			stat fs.FileInfo
		)

		if w.jsx != nil {
			buf, stat = w.processFile(before+tsxExt, true, true)
			if buf == nil {
				buf, stat = w.processFile(before+jsxExt, false, true)
			}
		}

		if buf == nil {
			buf, stat = w.processFile(before+tsExt, true, false)
		}

		if buf != nil {
			return &file{
				Reader:   bytes.NewReader(buf.Bytes()),
				name:     name,
				FileInfo: stat,
			}, nil
		}
	}

	return w.FS.Open(name)
}

func (w *wrapped) processFile(path string, isTS, isJSX bool) (*bytes.Buffer, fs.FileInfo) {
	f, err := w.FS.Open(path)
	if err != nil {
		return nil, nil
	}

	stat, err := f.Stat()
	if err != nil {
		return nil, nil
	}

	var buf bytes.Buffer

	t := parser.NewReaderTokeniser(f)

	var tk javascript.Tokeniser = &t

	if isTS {
		tk = javascript.AsTypescript(tk)
	}

	if isJSX {
		tk = javascript.AsJSX(tk)
	}

	m, err := javascript.ParseModule(tk)
	if err == nil && isJSX {
		err = jsx.Process(m, w.jsx)
	}

	if err == nil {
		fmt.Fprintf(&buf, "%#s", m)
	} else if w.errFn != nil {
		w.errFn(&buf, err)
	}

	return &buf, stat
}

type file struct {
	*bytes.Reader
	name string
	fs.FileInfo
}

func (file) Readdir(count int) ([]fs.FileInfo, error) {
	return nil, fs.ErrInvalid
}

func (f *file) Stat() (fs.FileInfo, error) {
	return f, nil
}

func (file) Close() error {
	return nil
}

func (f *file) Name() string {
	return f.name
}

func (f *file) Size() int64 {
	return f.Reader.Size()
}
