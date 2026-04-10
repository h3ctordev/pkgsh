package ui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestConfirmModal_Confirm(t *testing.T) {
	m := newConfirmModal("Desinstalar", "vlc, firefox")
	m2, confirmed, cancelled := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
	_ = m2
	if !confirmed {
		t.Error("expected confirmed=true on 's'")
	}
	if cancelled {
		t.Error("expected cancelled=false on 's'")
	}
}

func TestConfirmModal_Cancel(t *testing.T) {
	m := newConfirmModal("Desinstalar", "vlc")
	_, confirmed, cancelled := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	if confirmed {
		t.Error("expected confirmed=false on 'n'")
	}
	if !cancelled {
		t.Error("expected cancelled=true on 'n'")
	}
}

func TestConfirmModal_EscCancels(t *testing.T) {
	m := newConfirmModal("Desinstalar", "vlc")
	_, confirmed, cancelled := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if confirmed || !cancelled {
		t.Errorf("expected cancelled=true confirmed=false, got confirmed=%v cancelled=%v", confirmed, cancelled)
	}
}

func TestSudoModal_Input(t *testing.T) {
	m := newSudoModal()
	m2, _, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}})
	if m2.input != "p" {
		t.Errorf("expected input 'p', got %q", m2.input)
	}
}

func TestSudoModal_Backspace(t *testing.T) {
	m := newSudoModal()
	m, _, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	m, _, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'b'}})
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
