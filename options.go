package tsserver

import "io"

type Option func(*wrapped)

func ErrFn(errFn func(w io.Writer, err error)) Option {
	return func(wr *wrapped) {
		wr.errFn = errFn
	}
}
