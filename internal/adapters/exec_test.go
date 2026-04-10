package adapters_test

import (
	"bufio"
	"io"
	"strings"
	"testing"

	"github.com/hbustos/pkgsh/internal/adapters"
	"github.com/hbustos/pkgsh/internal/domain"
)

func TestRunCmd_OK(t *testing.T) {
	out, err := adapters.RunCmd([]string{"echo", "hello"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.TrimSpace(out) != "hello" {
		t.Fatalf("expected 'hello', got %q", out)
	}
}

func TestRunCmd_Error(t *testing.T) {
	_, err := adapters.RunCmd([]string{"false"})
	if err == nil {
		t.Fatal("expected error from 'false', got nil")
	}
}

func TestRunCmd_NoShellInjection(t *testing.T) {
	out, err := adapters.RunCmd([]string{"echo", "a;b"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.TrimSpace(out) != "a;b" {
		t.Fatalf("expected literal 'a;b', got %q", out)
	}
}

func TestStreamCmdStdin_PassesInputToProcess(t *testing.T) {
	pr, pw := io.Pipe()
	op := domain.NewOperation()
	go adapters.StreamCmdStdin([]string{"cat"}, pr, op.Writer())

	// pw.Write blocks until cat reads — deterministic, no sleep needed
	if _, err := pw.Write([]byte("pkgsh\n")); err != nil {
		t.Fatalf("failed to write to stdin pipe: %v", err)
	}
	pw.Close() // EOF → cat exits

	scanner := bufio.NewScanner(op.Reader())
	if !scanner.Scan() {
		t.Fatal("expected output from cat")
	}
	if scanner.Text() != "pkgsh" {
		t.Fatalf("expected 'pkgsh', got %q", scanner.Text())
	}
}
