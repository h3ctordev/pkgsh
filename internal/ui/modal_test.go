package ui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestModalHelp_RendersContent(t *testing.T) {
	m := newHelpModal()
	out := m.View(80)
	if !strings.Contains(out, "NAVEGACIÓN") {
		t.Errorf("help modal should contain NAVEGACIÓN, got: %q", out)
	}
	if !strings.Contains(out, "SELECCIÓN") {
		t.Errorf("help modal should contain SELECCIÓN, got: %q", out)
	}
}

func TestModalHelp_ClosesOnQuestionMark(t *testing.T) {
	m := newHelpModal()
	_, confirmed, cancelled := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("?")})
	if confirmed {
		t.Error("help modal should not confirm on ?")
	}
	if !cancelled {
		t.Error("help modal should cancel (close) on ?")
	}
}

func TestModalHelp_ClosesOnEsc(t *testing.T) {
	m := newHelpModal()
	_, _, cancelled := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if !cancelled {
		t.Error("help modal should cancel on Esc")
	}
}

func TestModalConfirm_Confirm(t *testing.T) {
	m := newConfirmModal("Test", "pkg1, pkg2")
	_, confirmed, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("s")})
	if !confirmed {
		t.Error("confirm modal should confirm on 's'")
	}
}
