package domain

import (
	"io"
	"sync"
)

type Operation struct {
	pr      *io.PipeReader
	pw      *io.PipeWriter
	stdinPR *io.PipeReader
	stdinPW *io.PipeWriter
	mu      sync.Mutex
	done    bool
	err     error
}

func NewOperation() *Operation {
	pr, pw := io.Pipe()
	spr, spw := io.Pipe()
	return &Operation{pr: pr, pw: pw, stdinPR: spr, stdinPW: spw}
}

// Stdin returns the stdin reader for the child process.
func (o *Operation) Stdin() io.Reader {
	return o.stdinPR
}

// SendInput writes a line of input to the operation's stdin.
func (o *Operation) SendInput(s string) {
	o.stdinPW.Write([]byte(s))
}

// CloseStdin closes the stdin pipe, signaling EOF to the child process.
func (o *Operation) CloseStdin() {
	o.stdinPW.Close()
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
