# tsserver

[![CI](https://github.com/MJKWoolnough/tsserver/actions/workflows/go-checks.yml/badge.svg)](https://github.com/MJKWoolnough/tsserver/actions)
[![Go Reference](https://pkg.go.dev/badge/vimagination.zapto.org/tsserver.svg)](https://pkg.go.dev/vimagination.zapto.org/tsserver)
[![Go Report Card](https://goreportcard.com/badge/vimagination.zapto.org/tsserver)](https://goreportcard.com/report/vimagination.zapto.org/tsserver)

--
    import "vimagination.zapto.org/tsserver"

Package tsserver implements a simple wrapper around fs.FS to intercept calls to open JavaScript files, and instead open their Typescript equivalents and generate the JavaScript, commenting out any Typescript parts.

## Highlights

 - Wrap any `fs.FS` implementation.
 - Automatically transpiles Typescript files to JavaScript. NB: Deprecated structure (e.g. namespaces, enums) will cause parsing errors, and unannotated type imports may cause import errors.
 - Customisable error output.

## Usage

```go
package main

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"strings"
	"time"

	"vimagination.zapto.org/tsserver"
)

type memFS map[string]string

func (m memFS) Open(path string) (fs.File, error) {
	if str, ok := m[path]; ok {
		return &memFile{
			name:   path,
			Reader: strings.NewReader(str),
			size:   len(str),
		}, nil
	}

	return nil, fs.ErrNotExist
}

type memFile struct {
	name string
	*strings.Reader
	size int
}

func (m *memFile) Stat() (fs.FileInfo, error) { return m, nil }
func (m *memFile) Close() error               { return nil }
func (m *memFile) Name() string               { return m.name }
func (m *memFile) Size() int64                { return int64(m.size) }
func (m *memFile) Mode() fs.FileMode          { return fs.ModePerm }
func (m *memFile) ModTime() time.Time         { return time.Now() }
func (m *memFile) IsDir() bool                { return false }
func (m *memFile) Sys() any                   { return m }

func main() {
	files := tsserver.WrapFS(memFS{
		"main.ts": "function hello(name: string) {console.log('Hello ' + name);}\n\nhello('Bob');",
	})

	file, err := files.Open("main.js")
	if err != nil {
		fmt.Println(err)

		return
	}
	defer file.Close()

	io.Copy(os.Stdout, file)

	// Output:
	// function hello(name/*: string*/) {console.log('Hello ' + name);}
	//
	// hello('Bob');
}
```

## Documentation

Full API docs can be found at:

https://pkg.go.dev/vimagination.zapto.org/tsserver
