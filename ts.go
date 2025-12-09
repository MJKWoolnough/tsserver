// Package tsserver implements a simple wrapper around fs.FS to intercept calls
// to open Javascript files, and instead open their Typescript equivalents and
// generate the Javascript.
package tsserver // import "vimagination.zapto.org/tsserver"

import (
	"bytes"
	"io"
	"io/fs"
	"strings"

	"vimagination.zapto.org/javascript"
	"vimagination.zapto.org/parser"
)

const (
	jsExt = ".js"
	tsExt = ".ts"
)

type wrapped struct {
	fs.FS
	errFn func(w io.Writer, err error)
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

func WrapFSWithErrorHandler(f fs.FS, errFn func(w io.Writer, err error)) fs.FS {
	return &wrapped{FS: f, errFn: errFn}
}

func (w *wrapped) Open(name string) (fs.File, error) {
	if strings.HasSuffix(name, jsExt) {
		if tsf, err := w.FS.Open(strings.TrimSuffix(name, jsExt) + tsExt); err == nil {
			if stat, err := tsf.Stat(); err == nil {
				tk := parser.NewReaderTokeniser(tsf)

				m, err := javascript.ParseModule(javascript.AsTypescript(&tk))
				if err == nil {
					var buf bytes.Buffer

					inTS := false

					for _, tk := range m.Tokens {
						if tk.IsTypescript() {
							if !inTS {
								inTS = true

								buf.WriteString("/*")
							}

							buf.WriteString(strings.ReplaceAll(tk.Data, "*/", "* /"))

							continue
						} else if inTS {
							buf.WriteString("*/")
							inTS = false
						}

						buf.WriteString(tk.Data)
					}

					return &file{
						Reader:   bytes.NewReader(buf.Bytes()),
						name:     name,
						FileInfo: stat,
					}, nil
				} else if w.errFn != nil {
					var buf bytes.Buffer

					w.errFn(&buf, err)

					return &file{
						Reader:   bytes.NewReader(buf.Bytes()),
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
