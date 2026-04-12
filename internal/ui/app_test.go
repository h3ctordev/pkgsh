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

func TestApp_SudoPrompt_OpensSudoModal(t *testing.T) {
	app := New(nil, nil, Options{})
	app.state.Operation = domain.NewOperation()

	model, _ := app.Update(operationLineMsg("PKGSH_SUDO:"))
	app = model.(AppModel)

	if app.modal == nil {
		t.Fatal("expected sudo modal to open on PKGSH_SUDO: line")
	}
	if app.modal.modalType != ModalSudo {
		t.Fatalf("expected ModalSudo, got %d", app.modal.modalType)
	}
	// Line should NOT be logged
	for _, line := range app.log.Lines() {
		if line == "PKGSH_SUDO:" {
			t.Fatal("PKGSH_SUDO: should not be logged")
		}
	}
}

func TestApp_SudoModal_Confirmed_SendsInputAndResumesStream(t *testing.T) {
	op := domain.NewOperation()

	// Read stdin in background — will block until SendInput writes
	received := make(chan string, 1)
	go func() {
		buf := make([]byte, 64)
		n, _ := op.StdinReader().Read(buf)
		received <- string(buf[:n])
	}()

	sudoModal := newSudoModal()
	sudoModal.input = "mypassword"

	app := New(nil, nil, Options{})
	app.state.Operation = op
	app.modal = &sudoModal

	// Press Enter to confirm
	model, cmd := app.Update(tea.KeyMsg{Type: tea.KeyEnter})
	app = model.(AppModel)

	if app.modal != nil {
		t.Fatal("expected modal to close after confirm")
	}
	if cmd == nil {
		t.Fatal("expected readLineCmd to be returned after sudo confirm")
	}

	// Verify the correct bytes were written to stdin.
	// The reader goroutine blocks until SendInput's goroutine writes; receive that first.
	if got := <-received; got != "mypassword\n" {
		t.Fatalf("expected stdin to receive %q, got %q", "mypassword\n", got)
	}
	op.Done(nil)
}

func TestApp_SudoModal_Cancelled_ClosesStdinAndResumesStream(t *testing.T) {
	op := domain.NewOperation()

	sudoModal := newSudoModal()
	app := New(nil, nil, Options{})
	app.state.Operation = op
	app.modal = &sudoModal

	// Press Esc to cancel
	model, cmd := app.Update(tea.KeyMsg{Type: tea.KeyEsc})
	app = model.(AppModel)

	if app.modal != nil {
		t.Fatal("expected modal to close after cancel")
	}
	if cmd == nil {
		t.Fatal("expected readLineCmd to be returned after sudo cancel")
	}
}

func TestApp_RegularLine_NotTreatedAsSudoPrompt(t *testing.T) {
	app := New(nil, nil, Options{})
	app.state.Operation = domain.NewOperation()

	model, _ := app.Update(operationLineMsg("Removing vim..."))
	app = model.(AppModel)

	if app.modal != nil {
		t.Fatal("expected no modal for regular line")
	}
	found := false
	for _, line := range app.log.Lines() {
		if line == "Removing vim..." {
			found = true
		}
	}
	if !found {
		t.Fatal("expected regular line to be logged")
	}
}

func TestApp_SecurityMode_BlocksSystemPackageInOperation(t *testing.T) {
	// bash es sistema; vim no lo es. Security mode activo.
	// Seleccionamos bash directamente en state.Packages vía campo privado selected.
	pkgs := []domain.Package{
		{Name: "vim", Manager: domain.ManagerApt, Version: "9.0"},
		{Name: "bash", Manager: domain.ManagerApt, Version: "5.2"},
	}
	adapters := map[domain.ManagerType]domain.PackageManager{
		domain.ManagerApt: &fakeAdapter{manager: domain.ManagerApt, output: "ok"},
	}
	app := New(pkgs, adapters, Options{SecurityMode: true})
	// bash está en state.Packages aunque no en Filtered (security mode lo oculta)
	// Forzar la selección directamente usando la clave "name:manager"
	app.list.selected["bash:apt"] = true
	app.currentKind = opRemove

	// Presionar 'd' con bash seleccionado → modal
	model, _ := app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	app = model.(AppModel)
	if app.modal == nil {
		t.Fatal("expected confirm modal")
	}

	// Confirmar con 'y' → bloqueo activo debe emitir [SECURITY]
	model, _ = app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})
	app = model.(AppModel)

	found := false
	for _, line := range app.log.Lines() {
		if strings.Contains(line, "[SECURITY]") && strings.Contains(line, "bash") {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected [SECURITY] line for bash, got %v", app.log.Lines())
	}
}

func TestApp_SecurityMode_AllBlockedNoOperationStarts(t *testing.T) {
	// Selección 100% de paquetes del sistema → no debe arrancar operación
	pkgs := []domain.Package{
		{Name: "bash", Manager: domain.ManagerApt, Version: "5.2"},
	}
	adapters := map[domain.ManagerType]domain.PackageManager{
		domain.ManagerApt: &fakeAdapter{manager: domain.ManagerApt, output: "ok"},
	}
	app := New(pkgs, adapters, Options{SecurityMode: true})
	app.list.selected["bash:apt"] = true
	app.currentKind = opRemove

	model, _ := app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	app = model.(AppModel)
	if app.modal == nil {
		t.Fatal("expected confirm modal")
	}
	model, _ = app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})
	app = model.(AppModel)

	if app.state.Operation != nil {
		t.Fatal("expected no Operation when all selected packages are system packages")
	}
}

func TestApp_SecurityModeToggle(t *testing.T) {
	pkgs := []domain.Package{
		{Name: "vim", Manager: domain.ManagerApt, Version: "9.0"},
		{Name: "bash", Manager: domain.ManagerApt, Version: "5.2"},   // sistema
		{Name: "libc6", Manager: domain.ManagerApt, Version: "2.35"}, // sistema
	}
	app := New(pkgs, nil, Options{SecurityMode: true})

	// Con security mode activo, solo vim en Filtered
	if len(app.state.Filtered) != 1 {
		t.Fatalf("expected 1 non-system package at start, got %d: %v", len(app.state.Filtered), app.state.Filtered)
	}

	// Toggle con 'S' → desactiva
	model, _ := app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'S'}})
	app = model.(AppModel)

	if app.state.SecurityMode {
		t.Fatal("expected SecurityMode=false after 'S' toggle")
	}
	if len(app.state.Filtered) != 3 {
		t.Fatalf("expected 3 packages after disabling security mode, got %d", len(app.state.Filtered))
	}

	// Toggle de nuevo con 'S' → activa
	model, _ = app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'S'}})
	app = model.(AppModel)

	if !app.state.SecurityMode {
		t.Fatal("expected SecurityMode=true after second 'S' toggle")
	}
	if len(app.state.Filtered) != 1 {
		t.Fatalf("expected 1 package after re-enabling security mode, got %d", len(app.state.Filtered))
	}
}

func TestApp_ProgressiveLoad_InitialState(t *testing.T) {
	adapters := map[domain.ManagerType]domain.PackageManager{
		domain.ManagerApt: &fakeAdapter{manager: domain.ManagerApt},
		domain.ManagerNpm: &fakeAdapter{manager: domain.ManagerNpm},
	}
	app := New(nil, adapters, Options{})

	if !app.loading {
		t.Fatal("expected loading=true when pkgs=nil and adapters provided")
	}
	if app.totalAdapters != 2 {
		t.Fatalf("expected totalAdapters=2, got %d", app.totalAdapters)
	}
	if app.loadedCount != 0 {
		t.Fatalf("expected loadedCount=0, got %d", app.loadedCount)
	}
}

func TestApp_ProgressiveLoad_AccumulatesAndFinishes(t *testing.T) {
	adapters := map[domain.ManagerType]domain.PackageManager{
		domain.ManagerApt: &fakeAdapter{manager: domain.ManagerApt},
		domain.ManagerNpm: &fakeAdapter{manager: domain.ManagerNpm},
	}
	app := New(nil, adapters, Options{})

	aptPkgs := []domain.Package{{Name: "vim", Manager: domain.ManagerApt, Version: "9.0"}}
	model, _ := app.Update(packagesLoadedMsg{manager: domain.ManagerApt, pkgs: aptPkgs})
	app = model.(AppModel)

	if !app.loading {
		t.Fatal("expected still loading after 1 of 2")
	}
	if len(app.state.Packages) != 1 {
		t.Fatalf("expected 1 accumulated package, got %d", len(app.state.Packages))
	}
	if len(app.state.Filtered) != 0 {
		t.Fatal("expected filtered empty until all loaded")
	}

	npmPkgs := []domain.Package{{Name: "lodash", Manager: domain.ManagerNpm, Version: "4.0"}}
	model, _ = app.Update(packagesLoadedMsg{manager: domain.ManagerNpm, pkgs: npmPkgs})
	app = model.(AppModel)

	if app.loading {
		t.Fatal("expected loading=false after all adapters loaded")
	}
	if len(app.state.Packages) != 2 {
		t.Fatalf("expected 2 packages, got %d", len(app.state.Packages))
	}
	if len(app.state.Filtered) != 2 {
		t.Fatalf("expected 2 filtered packages, got %d", len(app.state.Filtered))
	}
}

func TestApp_ViewFooter_ShowsProgressBar(t *testing.T) {
	adapters := map[domain.ManagerType]domain.PackageManager{
		domain.ManagerApt: &fakeAdapter{manager: domain.ManagerApt},
		domain.ManagerNpm: &fakeAdapter{manager: domain.ManagerNpm},
	}
	app := New(nil, adapters, Options{})
	app.width = 80

	footer := app.viewFooter(0)
	if !strings.Contains(footer, "Cargando") {
		t.Fatalf("expected 'Cargando' in footer during load, got: %q", footer)
	}
	if !strings.Contains(footer, "0/2") {
		t.Fatalf("expected '0/2' progress in footer, got: %q", footer)
	}

	model, _ := app.Update(packagesLoadedMsg{manager: domain.ManagerApt, pkgs: nil})
	app = model.(AppModel)

	footer = app.viewFooter(0)
	if !strings.Contains(footer, "1/2") {
		t.Fatalf("expected '1/2' progress after one load, got: %q", footer)
	}
}

func TestApp_ViewFooter_ShowsHintsWhenDone(t *testing.T) {
	app := makeApp(makePackages("vim"))
	app.width = 80

	footer := app.viewFooter(0)
	if strings.Contains(footer, "Cargando") {
		t.Fatalf("expected normal hints footer, got: %q", footer)
	}
	if !strings.Contains(footer, "Buscar") {
		t.Fatalf("expected hints in footer, got: %q", footer)
	}
}
