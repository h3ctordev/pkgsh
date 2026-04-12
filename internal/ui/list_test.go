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

// pkgTagsText devuelve los tags sin estilos ANSI para comparar en tests.
func pkgTagsText(p domain.Package) string {
	tags := ""
	if p.NewVersion != "" {
		tags += "UPD"
	}
	if domain.IsSystemPackage(p) {
		tags += "SYS"
	}
	if p.IsOrphan {
		tags += "ORP"
	}
	return tags
}

func TestListModel_JumpToFirst(t *testing.T) {
	lm := newListModel().SetItems(makePackages("a", "b", "c"))
	lm, _ = lm.Update(tea.KeyMsg{Type: tea.KeyDown})
	lm, _ = lm.Update(tea.KeyMsg{Type: tea.KeyDown})
	lm = lm.JumpToFirst()
	if lm.Cursor() != 0 {
		t.Fatalf("expected cursor 0 after JumpToFirst, got %d", lm.Cursor())
	}
}

func TestListModel_JumpToLast(t *testing.T) {
	lm := newListModel().SetItems(makePackages("a", "b", "c"))
	lm = lm.JumpToLast()
	if lm.Cursor() != 2 {
		t.Fatalf("expected cursor 2 after JumpToLast, got %d", lm.Cursor())
	}
}

func TestListModel_JumpToLast_Empty(t *testing.T) {
	lm := newListModel().SetItems(nil)
	lm = lm.JumpToLast() // no debe entrar en pánico
	if lm.Cursor() != 0 {
		t.Fatalf("expected cursor 0 on empty list, got %d", lm.Cursor())
	}
}

func TestListModel_JkNavigation(t *testing.T) {
	lm := newListModel().SetItems(makePackages("a", "b", "c"))
	lm, _ = lm.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	if lm.Cursor() != 1 {
		t.Fatalf("j should move down: expected 1, got %d", lm.Cursor())
	}
	lm, _ = lm.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")})
	if lm.Cursor() != 0 {
		t.Fatalf("k should move up: expected 0, got %d", lm.Cursor())
	}
}

func TestPkgTags(t *testing.T) {
	upd := domain.Package{Name: "x", Manager: domain.ManagerApt, NewVersion: "2.0"}
	if pkgTagsText(upd) == "" {
		t.Error("expected UPD tag for package with NewVersion")
	}
	orphan := domain.Package{Name: "x", Manager: domain.ManagerApt, IsOrphan: true}
	if pkgTagsText(orphan) == "" {
		t.Error("expected ORP tag for orphan package")
	}
	normal := domain.Package{Name: "x", Manager: domain.ManagerApt}
	if pkgTagsText(normal) != "" {
		t.Error("expected no tags for normal package")
	}
}
