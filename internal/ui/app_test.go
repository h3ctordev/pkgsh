package ui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/hbustos/pkgsh/internal/domain"
)

func makeApp(pkgs []domain.Package) AppModel {
	return New(pkgs, nil, Options{})
}

func TestApp_PanelSwitching(t *testing.T) {
	app := makeApp(makePackages("vim", "git"))

	if app.state.ActivePanel != domain.PanelList {
		t.Fatal("expected PanelList at start")
	}

	model, _ := app.Update(tea.KeyMsg{Type: tea.KeyTab})
	app = model.(AppModel)
	if app.state.ActivePanel != domain.PanelDetail {
		t.Fatalf("expected PanelDetail after Tab, got %d", app.state.ActivePanel)
	}

	model, _ = app.Update(tea.KeyMsg{Type: tea.KeyTab})
	app = model.(AppModel)
	if app.state.ActivePanel != domain.PanelLog {
		t.Fatalf("expected PanelLog after second Tab, got %d", app.state.ActivePanel)
	}

	model, _ = app.Update(tea.KeyMsg{Type: tea.KeyTab})
	app = model.(AppModel)
	if app.state.ActivePanel != domain.PanelList {
		t.Fatalf("expected PanelList wrap-around, got %d", app.state.ActivePanel)
	}
}

func TestApp_ConfirmModalOnDelete(t *testing.T) {
	app := makeApp(makePackages("vim", "git"))
	app.list = app.list.ToggleSelected() // seleccionar item en cursor 0

	model, _ := app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	app = model.(AppModel)

	if app.modal == nil {
		t.Fatal("expected confirm modal to open on 'd' with selection")
	}
	if app.modal.modalType != ModalConfirm {
		t.Fatalf("expected ModalConfirm, got %d", app.modal.modalType)
	}
}

func TestApp_NoModalOnDeleteWithoutSelection(t *testing.T) {
	app := makeApp(makePackages("vim"))
	model, _ := app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	app = model.(AppModel)
	if app.modal != nil {
		t.Fatal("expected no modal when nothing is selected")
	}
}

func TestApp_SearchActivation(t *testing.T) {
	app := makeApp(makePackages("vim", "git", "curl"))
	model, _ := app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	app = model.(AppModel)
	if !app.searching {
		t.Fatal("expected searching=true after '/'")
	}
}

func TestApp_ManagerFilterTab(t *testing.T) {
	pkgs := []domain.Package{
		{Name: "vim", Manager: domain.ManagerApt, Version: "9.0"},
		{Name: "node", Manager: domain.ManagerNpm, Version: "20.0"},
	}
	app := makeApp(pkgs)

	model, _ := app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'2'}})
	app = model.(AppModel)

	if app.state.ActiveTab != domain.ManagerApt {
		t.Fatalf("expected ManagerApt, got %s", app.state.ActiveTab)
	}
	if len(app.state.Filtered) != 1 || app.state.Filtered[0].Name != "vim" {
		t.Fatalf("expected [vim], got %v", app.state.Filtered)
	}
}

func TestApp_FilterAll(t *testing.T) {
	pkgs := []domain.Package{
		{Name: "vim", Manager: domain.ManagerApt, Version: "9.0"},
		{Name: "node", Manager: domain.ManagerNpm, Version: "20.0"},
	}
	app := makeApp(pkgs)
	// Filtrar a apt
	model, _ := app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'2'}})
	app = model.(AppModel)
	// Volver a todos
	model, _ = app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'1'}})
	app = model.(AppModel)

	if len(app.state.Filtered) != 2 {
		t.Fatalf("expected 2 packages after reset to All, got %d", len(app.state.Filtered))
	}
}

func TestApp_Quit(t *testing.T) {
	app := makeApp(makePackages("vim"))
	_, cmd := app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	if cmd == nil {
		t.Fatal("expected tea.Quit command on 'q'")
	}
}
