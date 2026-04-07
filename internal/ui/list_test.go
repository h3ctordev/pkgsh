package ui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/hbustos/pkgsh/internal/domain"
)

func makePackages(names ...string) []domain.Package {
	pkgs := make([]domain.Package, len(names))
	for i, n := range names {
		pkgs[i] = domain.Package{Name: n, Manager: domain.ManagerApt, Version: "1.0"}
	}
	return pkgs
}

func TestListModel_CursorMovement(t *testing.T) {
	lm := newListModel().SetItems(makePackages("a", "b", "c"))

	lm, _ = lm.Update(tea.KeyMsg{Type: tea.KeyDown})
	if lm.Cursor() != 1 {
		t.Fatalf("expected cursor 1, got %d", lm.Cursor())
	}

	lm, _ = lm.Update(tea.KeyMsg{Type: tea.KeyUp})
	if lm.Cursor() != 0 {
		t.Fatalf("expected cursor 0, got %d", lm.Cursor())
	}

	// no puede ir por debajo de 0
	lm, _ = lm.Update(tea.KeyMsg{Type: tea.KeyUp})
	if lm.Cursor() != 0 {
		t.Fatalf("cursor debe quedar en 0, got %d", lm.Cursor())
	}

	// no puede pasar del último
	lm, _ = lm.Update(tea.KeyMsg{Type: tea.KeyDown})
	lm, _ = lm.Update(tea.KeyMsg{Type: tea.KeyDown})
	lm, _ = lm.Update(tea.KeyMsg{Type: tea.KeyDown})
	if lm.Cursor() != 2 {
		t.Fatalf("cursor debe detenerse en 2, got %d", lm.Cursor())
	}
}

func TestListModel_Selection(t *testing.T) {
	lm := newListModel().SetItems(makePackages("a", "b", "c"))

	lm = lm.ToggleSelected()
	sel := lm.SelectedPackages()
	if len(sel) != 1 || sel[0].Name != "a" {
		t.Fatalf("expected [a], got %v", sel)
	}

	// toggle de nuevo debe deseleccionar
	lm = lm.ToggleSelected()
	if len(lm.SelectedPackages()) != 0 {
		t.Fatal("expected empty selection")
	}
}

func TestListModel_SelectAll(t *testing.T) {
	lm := newListModel().SetItems(makePackages("a", "b", "c"))
	lm = lm.SelectAll()
	if len(lm.SelectedPackages()) != 3 {
		t.Fatalf("expected 3 selected, got %d", len(lm.SelectedPackages()))
	}
}

func TestListModel_ClearSelection(t *testing.T) {
	lm := newListModel().SetItems(makePackages("a", "b", "c"))
	lm = lm.SelectAll().ClearSelection()
	if len(lm.SelectedPackages()) != 0 {
		t.Fatal("expected empty selection after clear")
	}
}

func TestListModel_SetItemsResetsCursor(t *testing.T) {
	lm := newListModel().SetItems(makePackages("a", "b", "c"))
	lm, _ = lm.Update(tea.KeyMsg{Type: tea.KeyDown})
	lm, _ = lm.Update(tea.KeyMsg{Type: tea.KeyDown})

	lm = lm.SetItems(makePackages("x"))
	if lm.Cursor() != 0 {
		t.Fatalf("cursor debe resetear a 0, got %d", lm.Cursor())
	}
}
