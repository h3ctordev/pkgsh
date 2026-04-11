package ui

import (
	"fmt"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/hbustos/pkgsh/internal/domain"
)

func makeApp(pkgs []domain.Package) AppModel {
	return New(pkgs, nil, Options{})
}

func TestApp_InitialState(t *testing.T) {
	pkgs := []domain.Package{
		{Name: "a", Version: "1.0", Manager: domain.ManagerApt},
		{Name: "b", Version: "2.0", Manager: domain.ManagerApt, NewVersion: "2.1"},
	}
	app := makeApp(pkgs)
	if len(app.state.Filtered) != 2 {
		t.Fatalf("expected 2 filtered packages, got %d", len(app.state.Filtered))
	}
}

func TestApp_FilterByManager(t *testing.T) {
	pkgs := []domain.Package{
		{Name: "a", Manager: domain.ManagerApt, Version: "1.0"},
		{Name: "b", Manager: domain.ManagerSnap, Version: "1.0"},
	}
	app := makeApp(pkgs)
	app.state.ActiveTab = domain.ManagerApt
	app = app.applyFilter()
	if len(app.state.Filtered) != 1 {
		t.Fatalf("expected 1 apt package, got %d", len(app.state.Filtered))
	}
}

func TestApp_SecurityModeKey(t *testing.T) {
	app := makeApp(nil)
	app.state.SecurityMode = false
	model, _ := app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("S")})
	app = model.(AppModel)
	if !app.state.SecurityMode {
		t.Error("S key should toggle security mode on")
	}
}

func TestApp_HelpModalKey(t *testing.T) {
	app := makeApp(nil)
	model, _ := app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("?")})
	app = model.(AppModel)
	if app.modal == nil {
		t.Error("? key should open help modal")
	}
	if app.modal.modalType != ModalHelp {
		t.Errorf("expected ModalHelp, got %v", app.modal.modalType)
	}
}

func TestApp_ReloadKey(t *testing.T) {
	app := makeApp(nil)
	app.state.Operation = nil
	_, cmd := app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("r")})
	if cmd == nil {
		t.Error("r key should return a reload command when no operation active")
	}
}

func TestApp_JkNavigation(t *testing.T) {
	pkgs := []domain.Package{
		{Name: "a", Manager: domain.ManagerApt, Version: "1.0"},
		{Name: "b", Manager: domain.ManagerApt, Version: "1.0"},
	}
	app := makeApp(pkgs)
	app.state.ActivePanel = domain.PanelList
	model, _ := app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	app = model.(AppModel)
	if app.list.Cursor() != 1 {
		t.Errorf("j should move cursor down: expected 1, got %d", app.list.Cursor())
	}
}

func TestApp_FooterWithSelection(t *testing.T) {
	pkgs := []domain.Package{
		{Name: "a", Manager: domain.ManagerApt, Version: "1.0"},
	}
	app := makeApp(pkgs)
	app.width = 120
	app.list = app.list.ToggleSelected()
	sel := app.list.AllSelected(app.state.Packages)
	footer := app.viewFooter(len(sel))
	if !strings.Contains(footer, "Desinstalar") {
		t.Errorf("footer should show Desinstalar when items selected, got: %q", footer)
	}
}

func TestApp_FooterWithoutSelection(t *testing.T) {
	app := makeApp(nil)
	app.width = 120
	footer := app.viewFooter(0)
	if strings.Contains(footer, "Desinstalar") {
		t.Errorf("footer should not show Desinstalar without selection, got: %q", footer)
	}
	if !strings.Contains(footer, "Ayuda") {
		t.Errorf("footer should always show Ayuda, got: %q", footer)
	}
}

func TestApp_LogCollapsedByDefault(t *testing.T) {
	app := makeApp(nil)
	if !app.logCollapsed {
		t.Error("log should be collapsed by default when no operation active")
	}
}

func TestApp_ViewRendersWithoutPanic(t *testing.T) {
	pkgs := []domain.Package{
		{Name: "a", Manager: domain.ManagerApt, Version: "1.0"},
	}
	app := makeApp(pkgs)
	app.width = 120
	app.height = 40
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("View() panicked: %v", r)
		}
	}()
	_ = app.View()
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
}

func TestApp_ConfirmModalOnDelete(t *testing.T) {
	app := makeApp(makePackages("vim", "git"))
	app.list = app.list.ToggleSelected()
	model, _ := app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	app = model.(AppModel)
	if app.modal == nil {
		t.Fatal("expected confirm modal to open on 'd' with selection")
	}
	if app.modal.modalType != ModalConfirm {
		t.Fatalf("expected ModalConfirm, got %d", app.modal.modalType)
	}
}

func TestApp_StartNextOp_SingleManager(t *testing.T) {
	pkgs := []domain.Package{
		{Name: "vim", Manager: domain.ManagerApt, Version: "1.0"},
	}
	adapters := map[domain.ManagerType]domain.PackageManager{
		domain.ManagerApt: &fakeAdapter{manager: domain.ManagerApt, output: "Removing vim"},
	}
	app := New(pkgs, adapters, Options{})
	app.pendingOps = []pendingOp{
		{manager: domain.ManagerApt, pkgs: pkgs},
	}
	app.currentKind = opRemove
	app, cmd := app.startNextOp()
	if app.state.Operation == nil {
		t.Fatal("expected Operation to be set after startNextOp")
	}
	if cmd == nil {
		t.Fatal("expected non-nil cmd from startNextOp")
	}
}

func TestApp_StartNextOp_SkipsMissingAdapter(t *testing.T) {
	pkgs := []domain.Package{{Name: "vim", Manager: domain.ManagerApt, Version: "1.0"}}
	app := New(pkgs, nil, Options{})
	app.pendingOps = []pendingOp{{manager: domain.ManagerApt, pkgs: pkgs}}
	app.currentKind = opRemove
	app, _ = app.startNextOp()
	found := false
	for _, line := range app.log.Lines() {
		if strings.Contains(line, "[SKIP]") {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected [SKIP] line when adapter is missing, got %v", app.log.Lines())
	}
}

func TestApp_OperationDone_LogsError(t *testing.T) {
	app := New(nil, nil, Options{})
	app.currentManager = domain.ManagerApt
	app.state.Operation = domain.NewOperation()
	app.pendingOps = []pendingOp{}
	model, _ := app.Update(operationDoneMsg{err: fmt.Errorf("permission denied")})
	app = model.(AppModel)
	found := false
	for _, line := range app.log.Lines() {
		if strings.Contains(line, "[ERROR]") && strings.Contains(line, "permission denied") {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected [ERROR] line in log, got %v", app.log.Lines())
	}
}

// fakeAdapter implementa domain.PackageManager con output configurable para tests.
type fakeAdapter struct {
	manager domain.ManagerType
	output  string
	err     error
}

func (f *fakeAdapter) Name() string { return string(f.manager) }
func (f *fakeAdapter) List() ([]domain.Package, error) { return nil, nil }
func (f *fakeAdapter) Info(pkg domain.Package) (domain.PackageInfo, error) {
	return domain.PackageInfo{Package: pkg}, nil
}
func (f *fakeAdapter) Remove(pkgs []domain.Package) *domain.Operation {
	op := domain.NewOperation()
	go func() {
		if f.output != "" {
			op.Writer().Write([]byte(f.output + "\n"))
		}
		op.Done(f.err)
	}()
	return op
}
func (f *fakeAdapter) Update(pkgs []domain.Package) *domain.Operation {
	return f.Remove(pkgs)
}
