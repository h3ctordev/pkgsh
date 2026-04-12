package domain

import (
	"io"
	"testing"
)

func TestOperation_StdinReader_ReceivesInput(t *testing.T) {
	op := NewOperation()
	op.SendInput("hello\n")

	buf := make([]byte, 32)
	n, err := op.StdinReader().Read(buf)
	if err != nil {
		t.Fatalf("unexpected error reading stdin: %v", err)
	}
	if string(buf[:n]) != "hello\n" {
		t.Fatalf("expected %q, got %q", "hello\n", string(buf[:n]))
	}
}

func TestOperation_CloseStdin_ReturnsEOF(t *testing.T) {
	op := NewOperation()
	op.CloseStdin()

	buf := make([]byte, 1)
	_, err := op.StdinReader().Read(buf)
	if err != io.ErrClosedPipe && err != io.EOF {
		t.Fatalf("expected EOF or ErrClosedPipe after CloseStdin, got: %v", err)
	}
}
