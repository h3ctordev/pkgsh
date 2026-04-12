package ui

import (
	"os"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestMain(m *testing.M) {
	// Force ANSI color output even in non-TTY test environments.
	os.Setenv("CLICOLOR_FORCE", "1")
	os.Exit(m.Run())
}

func TestLogModel_AppendLine(t *testing.T) {
	lm := newLogModel()
	lm = lm.appendLine("hello")
	lm = lm.appendLine("world")
	if len(lm.Lines()) != 2 {
		t.Fatalf("expected 2 lines, got %d", len(lm.Lines()))
	}
}

func TestLogModel_CRStrip(t *testing.T) {
	lm := newLogModel()
	lm = lm.appendLine("progress\rDone")
	if lm.Lines()[0] != "Done" {
		t.Errorf("expected carriage-return stripped line, got %q", lm.Lines()[0])
	}
}

func TestLogModel_AutoScroll(t *testing.T) {
	lm := newLogModel()
	lm = lm.appendLine("line1")
	if lm.scrollOff != 0 {
		t.Error("scrollOff should be 0 when autoScroll is true")
	}
}

func TestLogModel_ManualScroll(t *testing.T) {
	lm := newLogModel()
	lm, _ = lm.Update(tea.KeyMsg{Type: tea.KeyPgUp})
	if lm.autoScroll {
		t.Error("PgUp should disable autoScroll")
	}
}

func TestColorLine(t *testing.T) {
	tests := []struct {
		line     string
		contains string
	}{
		{"[ERROR] something failed", "\x1b"}, // has ANSI
		{"[SECURITY] blocked", "\x1b"},
		{"Removing firefox...", "\x1b"},
		{"normal log line", "\x1b"},
	}
	for _, tt := range tests {
		got := colorLine(tt.line)
		if !strings.Contains(got, tt.contains) {
			t.Errorf("colorLine(%q) should have ANSI codes, got %q", tt.line, got)
		}
	}
}

func TestLogModel_ViewCollapsed(t *testing.T) {
	lm := newLogModel()
	lm = lm.appendLine("some output")
	out := lm.View(80, 1, false, "")
	// height=1 collapsed: muestra mensaje sin panel completo
	if strings.Contains(out, "some output") {
		// en modo colapsado no debería mostrar líneas de log
		t.Error("collapsed log should not show log lines")
	}
}
