package tsserver

import (
	"io"
	"text/template"
)

type Option func(*wrapped)

func ErrFn(errFn func(w io.Writer, err error)) Option {
	return func(wr *wrapped) {
		wr.errFn = errFn
	}
}

func JSX(jsx *template.Template) Option {
	return func(wr *wrapped) {
		wr.jsx = jsx
	}
}
