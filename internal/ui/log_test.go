package ui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestLogAppendLine(t *testing.T) {
	lm := newLogModel()
	lm = lm.appendLine("primera línea")
	lm = lm.appendLine("segunda línea")
	if len(lm.lines) != 2 {
		t.Errorf("expected 2 lines, got %d", len(lm.lines))
	}
	if lm.lines[0] != "primera línea" {
		t.Errorf("unexpected first line: %s", lm.lines[0])
	}
}

func TestLogScrollUp(t *testing.T) {
	lm := newLogModel()
	for i := 0; i < 20; i++ {
		lm = lm.appendLine("línea")
	}
	lm.scrollOff = 5
	lm2, _ := lm.Update(tea.KeyMsg{Type: tea.KeyPgUp})
	if lm2.scrollOff >= lm.scrollOff {
		t.Errorf("expected scroll to decrease, got %d", lm2.scrollOff)
	}
}
