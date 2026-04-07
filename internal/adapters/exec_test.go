package adapters_test

import (
	"strings"
	"testing"

	"github.com/hbustos/pkgsh/internal/adapters"
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
