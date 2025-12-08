package tsserver_test

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

func Example() {
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
