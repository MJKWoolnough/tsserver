// Package tsserver implements a simple wrapper around fs.FS to intercept calls
// to open Javascript files, and instead open their Typescript equivelants and
// generate the Javascript.
package tsserver // import "vimagination.zapto.org/tsserver"

import (
	"bytes"
	"fmt"
	"io/fs"
	"strings"

	"vimagination.zapto.org/javascript"
	"vimagination.zapto.org/memio"
	"vimagination.zapto.org/parser"
)

const (
	jsExt = ".js"
	tsExt = ".ts"
)

type wrapped struct {
	fs.FS
}

// WrapFS takes a fs.FS and intercepts any calls to open .js files, and instead
// generates a file from a similarly named .ts file, if one exists.
//
// If a .ts file does not exists, fails to be converted to javascript, or if the
// file being opened is not a .js file then the file open will not be
// intercepted.
func WrapFS(f fs.FS) fs.FS {
	return &wrapped{FS: f}
}

func (w *wrapped) Open(name string) (fs.File, error) {
	if strings.HasSuffix(name, jsExt) {
		if tsf, err := w.FS.Open(strings.TrimSuffix(name, jsExt) + tsExt); err == nil {
			if stat, err := tsf.Stat(); err == nil {
				tk := parser.NewReaderTokeniser(tsf)
				m, err := javascript.ParseModule(javascript.AsTypescript(&tk))
				if err == nil {
					var buf memio.Buffer
					fmt.Fprintf(&buf, "%s", m)
					return &file{
						Reader:   bytes.NewReader(buf),
						name:     name,
						FileInfo: stat,
					}, nil
				}
			}
		}
	}
	return w.FS.Open(name)
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
