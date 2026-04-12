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

func TestModalConfirm_Cancel(t *testing.T) {
	m := newConfirmModal("Desinstalar", "vlc")
	_, confirmed, cancelled := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("n")})
	if confirmed {
		t.Error("expected confirmed=false on 'n'")
	}
	if !cancelled {
		t.Error("expected cancelled=true on 'n'")
	}
}

func TestModalConfirm_EscCancels(t *testing.T) {
	m := newConfirmModal("Desinstalar", "vlc")
	_, confirmed, cancelled := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if confirmed || !cancelled {
		t.Errorf("expected cancelled=true confirmed=false, got confirmed=%v cancelled=%v", confirmed, cancelled)
	}
}

func TestSudoModal_Input(t *testing.T) {
	m := newSudoModal()
	m2, _, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("p")})
	if m2.input != "p" {
		t.Errorf("expected input 'p', got %q", m2.input)
	}
}

func TestSudoModal_Backspace(t *testing.T) {
	m := newSudoModal()
	m, _, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("a")})
	m, _, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("b")})
	m, _, _ = m.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	if m.input != "a" {
		t.Errorf("expected 'a' after backspace, got %q", m.input)
	}
}

func TestSudoModal_ViewMasksWithAsterisk(t *testing.T) {
	m := newSudoModal()
	m, _, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
	m, _, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})

	view := m.View(80)
	if !strings.Contains(view, "**") {
		t.Fatalf("expected '**' mask in view, got: %q", view)
	}
	if strings.Contains(view, "••") {
		t.Fatal("expected no bullet mask '••', only '*'")
	}
}
