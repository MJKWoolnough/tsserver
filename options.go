package tsserver

import (
	"io"
	"text/template"
)

type Option func(*wrapped)

// ErrFn sets an error handler for the wrapped FS.
//
// Parsing and processing errors will be given to the supplied callback to be
// handled.
//
// Data written to the given writer will be in-place of the file requested.
func ErrFn(errFn func(w io.Writer, err error)) Option {
	return func(wr *wrapped) {
		wr.errFn = errFn
	}
}

// JSX enables JSX processing using the given template.
//
// Please refer to the JSX Process function docs for how to use the template.
// https://pkg.go.dev/vimagination.zapto.org/javascript/jsx#Process
func JSX(jsx *template.Template) Option {
	return func(wr *wrapped) {
		wr.jsx = jsx
	}
}
