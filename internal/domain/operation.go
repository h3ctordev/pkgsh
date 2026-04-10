package domain

import (
	"io"
	"sync"
)

type Operation struct {
	pr     *io.PipeReader
	pw     *io.PipeWriter
	stdinR *io.PipeReader
	stdinW *io.PipeWriter
	mu     sync.Mutex
	done   bool
	err    error
}

func NewOperation() *Operation {
	pr, pw := io.Pipe()
	sr, sw := io.Pipe()
	return &Operation{pr: pr, pw: pw, stdinR: sr, stdinW: sw}
}

func (o *Operation) Reader() io.Reader {
	return o.pr
}

func (o *Operation) Writer() io.WriteCloser {
	return o.pw
}

// StdinReader expone el extremo de lectura para cmd.Stdin.
func (o *Operation) StdinReader() io.Reader {
	return o.stdinR
}

// SendInput escribe s al stdin del proceso de forma no bloqueante.
func (o *Operation) SendInput(s string) {
	go o.stdinW.Write([]byte(s))
}

// CloseStdin cierra el stdin del proceso (EOF → sudo falla).
func (o *Operation) CloseStdin() {
	o.stdinW.Close()
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
