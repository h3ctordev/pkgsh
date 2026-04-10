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
	if len(app.pendingOps) != 0 {
		t.Fatalf("expected pendingOps to be empty, got %d", len(app.pendingOps))
	}
	if cmd == nil {
		t.Fatal("expected non-nil cmd (readLineCmd) from startNextOp")
	}
}

func TestApp_StartNextOp_EmptyQueue(t *testing.T) {
	app := New(nil, nil, Options{})
	app.pendingOps = []pendingOp{}

	app, _ = app.startNextOp()

	if app.state.Operation != nil {
		t.Fatal("expected nil Operation when queue is empty")
	}
	found := false
	for _, line := range app.log.Lines() {
		if line == "Listo." {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected 'Listo.' in log, got %v", app.log.Lines())
	}
}

func TestApp_StartNextOp_SkipsMissingAdapter(t *testing.T) {
	pkgs := []domain.Package{{Name: "vim", Manager: domain.ManagerApt, Version: "1.0"}}
	app := New(pkgs, nil, Options{}) // adapters = nil → ningún manager disponible
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

func TestApp_ConfirmDelete_StartsOperation(t *testing.T) {
	pkgs := []domain.Package{
		{Name: "vim", Manager: domain.ManagerApt, Version: "1.0"},
	}
	adapters := map[domain.ManagerType]domain.PackageManager{
		domain.ManagerApt: &fakeAdapter{manager: domain.ManagerApt, output: "ok"},
	}
	app := New(pkgs, adapters, Options{})
	// Seleccionar vim
	app.list = app.list.ToggleSelected()

	// Presionar 'd' → abre modal
	model, _ := app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	app = model.(AppModel)
	if app.modal == nil {
		t.Fatal("expected modal to open")
	}
	if app.currentKind != opRemove {
		t.Fatal("expected currentKind = opRemove")
	}

	// Confirmar con 'y'
	model, cmd := app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})
	app = model.(AppModel)

	if app.modal != nil {
		t.Fatal("expected modal to close after confirm")
	}
	if app.state.Operation == nil {
		t.Fatal("expected Operation to be set after confirm")
	}
	if cmd == nil {
		t.Fatal("expected readLineCmd to be returned")
	}
}

func TestApp_OperationDone_ChainsNext(t *testing.T) {
	pkgs1 := []domain.Package{{Name: "vim", Manager: domain.ManagerApt, Version: "1.0"}}
	pkgs2 := []domain.Package{{Name: "lodash", Manager: domain.ManagerNpm, Version: "4.0"}}
	adapterMap := map[domain.ManagerType]domain.PackageManager{
		domain.ManagerApt: &fakeAdapter{manager: domain.ManagerApt, output: "apt done"},
		domain.ManagerNpm: &fakeAdapter{manager: domain.ManagerNpm, output: "npm done"},
	}
	app := New(append(pkgs1, pkgs2...), adapterMap, Options{})
	// Simular que ya arrancó la primera op y queda una pendiente
	app.pendingOps = []pendingOp{{manager: domain.ManagerNpm, pkgs: pkgs2}}
	app.currentManager = domain.ManagerApt
	app.state.Operation = domain.NewOperation()

	// Simular que terminó sin error
	model, cmd := app.Update(operationDoneMsg{err: nil})
	app = model.(AppModel)

	// Debe haber arrancado la siguiente
	if app.state.Operation == nil {
		t.Fatal("expected second operation to start")
	}
	if cmd == nil {
		t.Fatal("expected readLineCmd for second operation")
	}
}

func TestApp_OperationDone_LogsError(t *testing.T) {
	app := New(nil, nil, Options{})
	app.currentManager = domain.ManagerApt
	app.state.Operation = domain.NewOperation()
	app.pendingOps = []pendingOp{} // cola vacía

	opErr := fmt.Errorf("permission denied")
	model, _ := app.Update(operationDoneMsg{err: opErr})
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
