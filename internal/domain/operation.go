package domain

import (
	"io"
	"sync"
)

type Operation struct {
	pr     *io.PipeReader
	pw     *io.PipeWriter
	mu     sync.Mutex
	done   bool
	err    error
}

func NewOperation() *Operation {
	pr, pw := io.Pipe()
	return &Operation{pr: pr, pw: pw}
}

func (o *Operation) Reader() io.Reader {
	return o.pr
}

func (o *Operation) Writer() io.WriteCloser {
	return o.pw
}

func (o *Operation) Done(err error) {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.done = true
	o.err = err
	o.pw.CloseWithError(err)
}

func (o *Operation) Err() error {
	o.mu.Lock()
	defer o.mu.Unlock()
	return o.err
}
