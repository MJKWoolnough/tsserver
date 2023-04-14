package tsserver // import "vimagination.zapto.org/tsserver"

import (
	"bytes"
	"fmt"
	"io/fs"
	"net/http"
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
	http.FileSystem
}

func WrapFileSystem(fs http.FileSystem) http.FileSystem {
	return &wrapped{FileSystem: fs}
}

func (w *wrapped) Open(name string) (http.File, error) {
	if strings.HasSuffix(name, jsExt) {
		if tsf, err := w.FileSystem.Open(strings.TrimSuffix(name, jsExt) + tsExt); err == nil {
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
	return w.FileSystem.Open(name)
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
