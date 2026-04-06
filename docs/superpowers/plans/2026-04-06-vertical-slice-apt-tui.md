# pkgsh Vertical Slice — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build the first working vertical slice of pkgsh: full Bubble Tea TUI with mock data (PR 1), real apt adapter with tests (PR 2), and wire them together (PR 3).

**Architecture:** Three branches merged sequentially to master. PR 1 establishes the complete TUI layout (tabs + list + detail + log + modal) with 12 hardcoded mock packages and simulated operations. PR 2 implements the apt adapter in isolation with unit tests against fixtures and an integration test. PR 3 replaces the mock slice in `main.go` with real goroutine-based adapter calls.

**Tech Stack:** Go 1.22, Bubble Tea v0.26.6, Bubbles v0.18.0, Lipgloss v0.11.0, `exec.Cmd` for system commands (always as `[]string`, never shell interpolation).

---

## File Map

| File | Action | Responsibility |
|---|---|---|
| `internal/domain/filter.go` | Modify | Add `Sort()` function |
| `internal/ui/app.go` | Create | Root `tea.Model`, layout, keybindings, operation lifecycle |
| `internal/ui/log.go` | Create | Log panel model, `readLineCmd`, streaming |
| `internal/ui/modal.go` | Create | Confirm modal + sudo prompt modal |
| `cmd/pkgsh/main.go` | Modify | Launch TUI with mock (PR1) then real adapter (PR3) |
| `internal/adapters/apt/adapter.go` | Modify | Implement `List()`, `Remove()`, `Update()`, `Info()` |
| `internal/adapters/apt/adapter_test.go` | Create | Fixture + integration tests |

---

## PR 1 — TUI con datos mock

**Branch:** `feat/tui-mock`

---

### Task 1: Crear rama y agregar dependencias de Bubble Tea

**Files:**
- Modify: `go.mod`, `go.sum`

- [ ] **Step 1: Crear rama**

```bash
git checkout master
git checkout -b feat/tui-mock
```

Expected: `Switched to a new branch 'feat/tui-mock'`

- [ ] **Step 2: Agregar dependencias**

```bash
go get github.com/charmbracelet/bubbletea@v0.26.6
go get github.com/charmbracelet/bubbles@v0.18.0
go get github.com/charmbracelet/lipgloss@v0.11.0
go mod tidy
```

Expected: `go: added github.com/charmbracelet/bubbletea v0.26.6` (y similares para los otros dos)

- [ ] **Step 3: Verificar que el proyecto compila**

```bash
go build ./...
```

Expected: sin output (sin errores)

- [ ] **Step 4: Commit**

```bash
git add go.mod go.sum
git commit -m "chore(deps): add bubbletea, bubbles, lipgloss"
```

---

### Task 2: Agregar Sort() al dominio

**Files:**
- Modify: `internal/domain/filter.go`
- Test: `internal/domain/filter_test.go`

- [ ] **Step 1: Escribir el test que falla**

Crear `internal/domain/filter_test.go`:

```go
package domain_test

import (
	"testing"

	"github.com/hbustos/pkgsh/internal/domain"
)

func testPkgs() []domain.Package {
	return []domain.Package{
		{Name: "zsh",     Manager: domain.ManagerApt,  Version: "5.9",    Size: 2_900_000},
		{Name: "apt",     Manager: domain.ManagerApt,  Version: "2.6.1",  Size: 4_096_000},
		{Name: "node",    Manager: domain.ManagerNpm,  Version: "20.1.0", Size: 98_000_000},
		{Name: "firefox", Manager: domain.ManagerApt,  Version: "126.0",  Size: 245_000_000},
	}
}

func TestSortByName(t *testing.T) {
	sorted := domain.Sort(testPkgs(), domain.SortByName)
	if sorted[0].Name != "apt" || sorted[3].Name != "zsh" {
		t.Errorf("expected alphabetical order, got %v", namesOf(sorted))
	}
}

func TestSortBySize(t *testing.T) {
	sorted := domain.Sort(testPkgs(), domain.SortBySize)
	if sorted[0].Name != "firefox" {
		t.Errorf("expected largest first, got %s", sorted[0].Name)
	}
}

func TestSortByManager(t *testing.T) {
	sorted := domain.Sort(testPkgs(), domain.SortByManager)
	// apt < npm alphabetically
	if sorted[len(sorted)-1].Manager != domain.ManagerNpm {
		t.Errorf("expected npm last, got %s", sorted[len(sorted)-1].Manager)
	}
}

func TestFilterByManager(t *testing.T) {
	filtered := domain.Filter(testPkgs(), "", domain.ManagerNpm)
	if len(filtered) != 1 || filtered[0].Name != "node" {
		t.Errorf("expected 1 npm package, got %v", filtered)
	}
}

func TestFilterByQuery(t *testing.T) {
	filtered := domain.Filter(testPkgs(), "fire", "")
	if len(filtered) != 1 || filtered[0].Name != "firefox" {
		t.Errorf("expected firefox, got %v", filtered)
	}
}

func namesOf(pkgs []domain.Package) []string {
	out := make([]string, len(pkgs))
	for i, p := range pkgs {
		out[i] = p.Name
	}
	return out
}
```

- [ ] **Step 2: Verificar que falla**

```bash
go test ./internal/domain/...
```

Expected: `FAIL — undefined: domain.Sort`

- [ ] **Step 3: Implementar Sort()**

Agregar al final de `internal/domain/filter.go`:

```go
import (
	"sort"
	"strings"
)

func Sort(pkgs []Package, by SortField) []Package {
	sorted := make([]Package, len(pkgs))
	copy(sorted, pkgs)
	sort.SliceStable(sorted, func(i, j int) bool {
		switch by {
		case SortByManager:
			return string(sorted[i].Manager) < string(sorted[j].Manager)
		case SortByVersion:
			return sorted[i].Version < sorted[j].Version
		case SortBySize:
			return sorted[i].Size > sorted[j].Size
		default: // SortByName
			return strings.ToLower(sorted[i].Name) < strings.ToLower(sorted[j].Name)
		}
	})
	return sorted
}
```

> **Nota:** `filter.go` ya importa `"strings"`. Solo agrega `"sort"` al bloque de imports existente.

- [ ] **Step 4: Verificar que los tests pasan**

```bash
go test ./internal/domain/...
```

Expected: `ok  github.com/hbustos/pkgsh/internal/domain`

- [ ] **Step 5: Commit**

```bash
git add internal/domain/filter.go internal/domain/filter_test.go
git commit -m "feat(domain): add Sort() function with SliceStable"
```

---

### Task 3: Crear internal/ui/log.go

**Files:**
- Create: `internal/ui/log.go`
- Test: `internal/ui/log_test.go`

- [ ] **Step 1: Escribir el test que falla**

Crear `internal/ui/log_test.go`:

```go
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
```

- [ ] **Step 2: Verificar que falla**

```bash
go test ./internal/ui/...
```

Expected: `FAIL — cannot find package` (el directorio no existe aún, lo cual está bien — el test fallará al crear el package)

- [ ] **Step 3: Crear internal/ui/log.go**

```go
package ui

import (
	"bufio"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/hbustos/pkgsh/internal/domain"
)

// operationLineMsg llega cada vez que se lee una línea de la operación activa.
type operationLineMsg string

// operationDoneMsg llega cuando la operación activa termina (err puede ser nil).
type operationDoneMsg struct{ err error }

type LogModel struct {
	lines     []string
	scrollOff int
}

func newLogModel() LogModel {
	return LogModel{}
}

func (lm LogModel) appendLine(line string) LogModel {
	lm.lines = append(lm.lines, line)
	return lm
}

// Update maneja PgUp/PgDn cuando el panel de log está activo.
func (lm LogModel) Update(msg tea.Msg) (LogModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyPgUp:
			if lm.scrollOff > 0 {
				lm.scrollOff -= 5
				if lm.scrollOff < 0 {
					lm.scrollOff = 0
				}
			}
		case tea.KeyPgDown:
			lm.scrollOff += 5
		}
	}
	return lm, nil
}

// View renderiza el panel de log mostrando las últimas líneas visibles.
func (lm LogModel) View(width, height int, active bool) string {
	style := logPanelStyle(active, width, height)
	if len(lm.lines) == 0 {
		return style.Render(lipgloss.NewStyle().Faint(true).Render("Sin operaciones"))
	}

	// Calcular ventana de líneas visibles
	start := len(lm.lines) - height - lm.scrollOff
	if start < 0 {
		start = 0
	}
	end := start + height
	if end > len(lm.lines) {
		end = len(lm.lines)
	}

	var rendered []string
	for _, line := range lm.lines[start:end] {
		rendered = append(rendered, line)
	}
	return style.Render(lipgloss.JoinVertical(lipgloss.Left, rendered...))
}

// readLineCmd devuelve un tea.Cmd que lee la próxima línea de la operación activa.
// Se debe llamar recursivamente hasta recibir operationDoneMsg.
func readLineCmd(op *domain.Operation) tea.Cmd {
	return func() tea.Msg {
		scanner := bufio.NewScanner(op.Reader())
		if scanner.Scan() {
			return operationLineMsg(scanner.Text())
		}
		return operationDoneMsg{err: scanner.Err()}
	}
}

func logPanelStyle(active bool, width, height int) lipgloss.Style {
	s := lipgloss.NewStyle().
		Width(width - 2).
		Height(height).
		Border(lipgloss.RoundedBorder()).
		Padding(0, 1)
	if active {
		s = s.BorderForeground(lipgloss.Color("86"))
	}
	return s
}
```

- [ ] **Step 4: Verificar que los tests pasan**

```bash
go test ./internal/ui/...
```

Expected: `ok  github.com/hbustos/pkgsh/internal/ui`

- [ ] **Step 5: Commit**

```bash
git add internal/ui/log.go internal/ui/log_test.go
git commit -m "feat(ui): add LogModel with streaming line reader"
```

---

### Task 4: Crear internal/ui/modal.go

**Files:**
- Create: `internal/ui/modal.go`
- Test: `internal/ui/modal_test.go`

- [ ] **Step 1: Escribir el test que falla**

Crear `internal/ui/modal_test.go`:

```go
package ui

import (
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
```

- [ ] **Step 2: Verificar que falla**

```bash
go test ./internal/ui/...
```

Expected: `FAIL — undefined: newConfirmModal`

- [ ] **Step 3: Crear internal/ui/modal.go**

```go
package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type ModalType int

const (
	ModalConfirm ModalType = iota
	ModalSudo
	ModalQuitConfirm
)

type ModalModel struct {
	modalType ModalType
	title     string
	body      string
	input     string
}

func newConfirmModal(title, pkgNames string) ModalModel {
	return ModalModel{
		modalType: ModalConfirm,
		title:     title,
		body:      pkgNames,
	}
}

func newSudoModal() ModalModel {
	return ModalModel{modalType: ModalSudo}
}

func newQuitConfirmModal() ModalModel {
	return ModalModel{modalType: ModalQuitConfirm, title: "Operación en curso"}
}

// Update procesa keypresses del modal.
// Retorna (modelo actualizado, confirmado, cancelado).
func (m ModalModel) Update(msg tea.Msg) (ModalModel, bool, bool) {
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, false, false
	}

	switch m.modalType {
	case ModalConfirm, ModalQuitConfirm:
		switch keyMsg.String() {
		case "s", "y":
			return m, true, false
		case "n", "esc":
			return m, false, true
		}

	case ModalSudo:
		switch keyMsg.Type {
		case tea.KeyEnter:
			return m, true, false
		case tea.KeyEsc:
			m.input = ""
			return m, false, true
		case tea.KeyBackspace:
			if len(m.input) > 0 {
				m.input = m.input[:len(m.input)-1]
			}
		default:
			if keyMsg.Type == tea.KeyRunes {
				m.input += string(keyMsg.Runes)
			}
		}
	}

	return m, false, false
}

// View renderiza el modal centrado en la pantalla.
func (m ModalModel) View(width int) string {
	var content string
	switch m.modalType {
	case ModalConfirm:
		content = fmt.Sprintf(
			"%s\n\n%s\n\n[s] Confirmar    [n/Esc] Cancelar",
			m.title, m.body,
		)
	case ModalQuitConfirm:
		content = fmt.Sprintf(
			"%s\n\n¿Salir de todas formas?\n\n[s] Salir    [n/Esc] Cancelar",
			m.title,
		)
	case ModalSudo:
		masked := strings.Repeat("•", len(m.input))
		content = fmt.Sprintf(
			"Se requiere contraseña de sudo:\n\n> %s\n\n[Enter] Confirmar    [Esc] Cancelar",
			masked,
		)
	}

	boxWidth := 50
	if boxWidth > width-4 {
		boxWidth = width - 4
	}

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("212")).
		Padding(1, 2).
		Width(boxWidth).
		Render(content)

	return lipgloss.Place(width, 10, lipgloss.Center, lipgloss.Center, box)
}
```

- [ ] **Step 4: Verificar que los tests pasan**

```bash
go test ./internal/ui/...
```

Expected: `ok  github.com/hbustos/pkgsh/internal/ui`

- [ ] **Step 5: Commit**

```bash
git add internal/ui/modal.go internal/ui/modal_test.go
git commit -m "feat(ui): add ModalModel for confirm and sudo prompts"
```

---

### Task 5: Crear internal/ui/app.go

**Files:**
- Create: `internal/ui/app.go`
- Test: `internal/ui/app_test.go`

- [ ] **Step 1: Escribir tests que fallan**

Crear `internal/ui/app_test.go`:

```go
package ui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/hbustos/pkgsh/internal/domain"
)

func testPackages() []domain.Package {
	return []domain.Package{
		{Name: "apt",     Version: "2.6.1",   Manager: domain.ManagerApt,  Size: 4_096_000,   Description: "Advanced Package Tool",       IsNative: true,  Origin: "ubuntu"},
		{Name: "bash",    Version: "5.2.15",  Manager: domain.ManagerApt,  Size: 1_800_000,   Description: "GNU Bourne Again SHell",       IsNative: true,  Origin: "ubuntu"},
		{Name: "firefox", Version: "126.0.1", NewVersion: "127.0", Manager: domain.ManagerApt, Size: 245_000_000, Description: "Web browser from Mozilla", IsNative: true, Origin: "ubuntu"},
		{Name: "node",    Version: "20.11.0", Manager: domain.ManagerNpm,  Size: 98_000_000,  Description: "JavaScript runtime (global)",  IsNative: false, Origin: "npmjs"},
		{Name: "snapd",   Version: "2.63",    Manager: domain.ManagerSnap, Size: 87_000_000,  Description: "Snap daemon and tooling",      IsNative: false, Origin: "snap-store"},
	}
}

func TestNew_InitializesFiltered(t *testing.T) {
	m := New(testPackages())
	if len(m.state.Filtered) != len(testPackages()) {
		t.Errorf("expected Filtered=%d, got %d", len(testPackages()), len(m.state.Filtered))
	}
}

func TestCursorDown(t *testing.T) {
	m := New(testPackages())
	m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	if m2.(Model).cursor != 1 {
		t.Errorf("expected cursor 1, got %d", m2.(Model).cursor)
	}
}

func TestCursorClampsAtBottom(t *testing.T) {
	m := New(testPackages())
	m.cursor = len(testPackages()) - 1
	m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	if m2.(Model).cursor != len(testPackages())-1 {
		t.Errorf("expected cursor to stay at %d, got %d", len(testPackages())-1, m2.(Model).cursor)
	}
}

func TestSpaceTogglesSelection(t *testing.T) {
	m := New(testPackages())
	m2, _ := m.Update(tea.KeyMsg{Type: tea.KeySpace})
	m3 := m2.(Model)
	if !m3.state.Selected[0] {
		t.Error("expected index 0 to be selected after Space")
	}
	m4, _ := m3.Update(tea.KeyMsg{Type: tea.KeySpace})
	if m4.(Model).state.Selected[0] {
		t.Error("expected index 0 to be deselected after second Space")
	}
}

func TestTabFilterByApt(t *testing.T) {
	m := New(testPackages())
	m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'2'}})
	m3 := m2.(Model)
	for _, p := range m3.state.Filtered {
		if p.Manager != domain.ManagerApt {
			t.Errorf("expected only apt packages after pressing '2', got %s", p.Manager)
		}
	}
}

func TestSearchFilters(t *testing.T) {
	m := New(testPackages())
	// Activar búsqueda
	m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	// Escribir 'fir'
	for _, r := range "fir" {
		m2, _ = m2.(Model).Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	m3 := m2.(Model)
	if len(m3.state.Filtered) != 1 || m3.state.Filtered[0].Name != "firefox" {
		t.Errorf("expected firefox after searching 'fir', got %v", m3.state.Filtered)
	}
}
```

- [ ] **Step 2: Verificar que falla**

```bash
go test ./internal/ui/...
```

Expected: `FAIL — undefined: New`

- [ ] **Step 3: Crear internal/ui/app.go**

```go
package ui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/hbustos/pkgsh/internal/domain"
)

// tickMsg se usa para forzar re-render periódico durante operaciones.
type tickMsg struct{}

type Model struct {
	state     domain.AppState
	cursor    int
	width     int
	height    int
	log       LogModel
	modal     ModalModel
	showModal bool
	searching bool
}

func New(pkgs []domain.Package) Model {
	state := domain.AppState{
		Packages: pkgs,
		Filtered: domain.Sort(pkgs, domain.SortByName),
		Selected: make(map[int]bool),
		SortBy:   domain.SortByName,
	}
	return Model{
		state:  state,
		log:    newLogModel(),
		width:  80,
		height: 24,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Modal intercepta todos los keys cuando está activo
	if m.showModal {
		keyMsg, isKey := msg.(tea.KeyMsg)
		if isKey {
			var confirmed, cancelled bool
			m.modal, confirmed, cancelled = m.modal.Update(keyMsg)
			if confirmed {
				return m.handleModalConfirm()
			}
			if cancelled {
				m.showModal = false
			}
		}
		return m, nil
	}

	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case operationLineMsg:
		m.log = m.log.appendLine(string(msg))
		if m.state.Operation != nil {
			return m, readLineCmd(m.state.Operation)
		}

	case operationDoneMsg:
		if msg.err != nil {
			m.log = m.log.appendLine("ERROR: " + msg.err.Error())
		}
		m.state.Operation = nil
		m.state.Selected = make(map[int]bool)

	case tickMsg:
		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg)
	}

	return m, nil
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Modo búsqueda: redirige teclas al buffer de búsqueda
	if m.searching {
		switch msg.Type {
		case tea.KeyEsc, tea.KeyEnter:
			m.searching = false
		case tea.KeyBackspace:
			if len(m.state.SearchQuery) > 0 {
				m.state.SearchQuery = m.state.SearchQuery[:len(m.state.SearchQuery)-1]
				m.applyFilter()
			}
		default:
			if msg.Type == tea.KeyRunes {
				m.state.SearchQuery += string(msg.Runes)
				m.applyFilter()
				m.cursor = 0
			}
		}
		return m, nil
	}

	switch msg.String() {
	case "q", "ctrl+c":
		if m.state.Operation != nil {
			m.modal = newQuitConfirmModal()
			m.showModal = true
			return m, nil
		}
		return m, tea.Quit

	case "tab":
		m.state.ActivePanel = (m.state.ActivePanel + 1) % 3

	case "up":
		if m.state.ActivePanel == domain.PanelLog {
			m.log, _ = m.log.Update(msg)
		} else if m.cursor > 0 {
			m.cursor--
		}

	case "down":
		if m.state.ActivePanel == domain.PanelLog {
			m.log, _ = m.log.Update(msg)
		} else if m.cursor < len(m.state.Filtered)-1 {
			m.cursor++
		}

	case "pgup":
		m.log, _ = m.log.Update(msg)

	case "pgdown":
		m.log, _ = m.log.Update(msg)

	case " ":
		m.state.Selected[m.cursor] = !m.state.Selected[m.cursor]

	case "a":
		for i := range m.state.Filtered {
			m.state.Selected[i] = true
		}

	case "esc":
		m.state.Selected = make(map[int]bool)
		m.state.SearchQuery = ""
		m.applyFilter()
		m.cursor = 0

	case "/":
		m.searching = true

	case "s":
		m.state.SortBy = (m.state.SortBy + 1) % 4
		m.applyFilter()
		m.cursor = 0

	case "1":
		m.setTab("")
	case "2":
		m.setTab(domain.ManagerApt)
	case "3":
		m.setTab(domain.ManagerSnap)
	case "4":
		m.setTab(domain.ManagerFlatpak)
	case "5":
		m.setTab(domain.ManagerDpkg)
	case "6":
		m.setTab(domain.ManagerPip)
	case "7":
		m.setTab(domain.ManagerNpm)
	case "8":
		m.setTab(domain.ManagerAppImage)

	case "d":
		return m.startOperation("Desinstalar")

	case "u":
		return m.startOperation("Actualizar")
	}

	return m, nil
}

func (m *Model) setTab(manager domain.ManagerType) {
	m.state.ActiveTab = manager
	m.state.Selected = make(map[int]bool)
	m.applyFilter()
	m.cursor = 0
}

func (m *Model) applyFilter() {
	filtered := domain.Filter(m.state.Packages, m.state.SearchQuery, m.state.ActiveTab)
	m.state.Filtered = domain.Sort(filtered, m.state.SortBy)
}

func (m Model) selectedPackages() []domain.Package {
	if len(m.state.Selected) == 0 {
		if m.cursor < len(m.state.Filtered) {
			return []domain.Package{m.state.Filtered[m.cursor]}
		}
		return nil
	}
	var pkgs []domain.Package
	for idx := range m.state.Selected {
		if m.state.Selected[idx] && idx < len(m.state.Filtered) {
			pkgs = append(pkgs, m.state.Filtered[idx])
		}
	}
	return pkgs
}

func (m Model) startOperation(action string) (tea.Model, tea.Cmd) {
	pkgs := m.selectedPackages()
	if len(pkgs) == 0 {
		return m, nil
	}
	names := make([]string, len(pkgs))
	for i, p := range pkgs {
		names[i] = p.Name
	}
	body := strings.Join(names, ", ")
	m.modal = newConfirmModal(
		fmt.Sprintf("%s %d paquete(s):", action, len(pkgs)),
		body,
	)
	m.showModal = true
	m.state.ActivePanel = domain.PanelLog
	return m, nil
}

func (m Model) handleModalConfirm() (tea.Model, tea.Cmd) {
	m.showModal = false
	pkgs := m.selectedPackages()
	if len(pkgs) == 0 {
		return m, nil
	}

	// Operación mock: escribe líneas simuladas al pipe
	op := domain.NewOperation()
	m.state.Operation = op
	m.log = newLogModel()

	go func() {
		w := op.Writer()
		for _, p := range pkgs {
			time.Sleep(300 * time.Millisecond)
			fmt.Fprintf(w, "Procesando %s... listo.\n", p.Name)
		}
		op.Done(nil)
	}()

	return m, readLineCmd(op)
}

// View ensambla el layout completo.
func (m Model) View() string {
	if m.width == 0 {
		return "Cargando..."
	}

	tabBar := m.renderTabBar()
	helpBar := m.renderHelpBar()

	logH := 8
	contentH := m.height - 3 - logH // 3 = tabBar + helpBar + margin
	if contentH < 4 {
		contentH = 4
	}

	listW := m.width / 2
	detailW := m.width - listW

	listView := m.renderList(listW, contentH)
	detailView := m.renderDetail(detailW, contentH)
	content := lipgloss.JoinHorizontal(lipgloss.Top, listView, detailView)

	logView := m.log.View(m.width, logH, m.state.ActivePanel == domain.PanelLog)

	layout := lipgloss.JoinVertical(lipgloss.Left,
		tabBar,
		content,
		logView,
		helpBar,
	)

	if m.showModal {
		// Renderizar modal sobre el layout
		overlay := m.modal.View(m.width)
		_ = overlay
		// Simplificado: mostrar modal debajo del layout en PR1
		return lipgloss.JoinVertical(lipgloss.Left, layout, overlay)
	}

	return layout
}

func (m Model) renderTabBar() string {
	managers := []struct {
		key  string
		name string
		mgr  domain.ManagerType
	}{
		{"1", "Todos", ""},
		{"2", "apt", domain.ManagerApt},
		{"3", "snap", domain.ManagerSnap},
		{"4", "flatpak", domain.ManagerFlatpak},
		{"5", "dpkg", domain.ManagerDpkg},
		{"6", "pip", domain.ManagerPip},
		{"7", "npm", domain.ManagerNpm},
		{"8", "AppImage", domain.ManagerAppImage},
	}

	base := lipgloss.NewStyle().Padding(0, 1)
	active := base.Bold(true).Foreground(lipgloss.Color("86")).Underline(true)

	var tabs []string
	for _, t := range managers {
		label := fmt.Sprintf("[%s] %s", t.key, t.name)
		if m.state.ActiveTab == t.mgr {
			tabs = append(tabs, active.Render(label))
		} else {
			tabs = append(tabs, base.Render(label))
		}
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, tabs...)
}

func (m Model) renderList(width, height int) string {
	active := m.state.ActivePanel == domain.PanelList
	border := lipgloss.NewStyle().
		Width(width - 2).
		Height(height).
		Border(lipgloss.RoundedBorder()).
		Padding(0, 1)
	if active {
		border = border.BorderForeground(lipgloss.Color("86"))
	}

	header := lipgloss.NewStyle().Faint(true).Render(
		fmt.Sprintf("  %-28s %-8s %s", "Paquete", "Gestor", "Versión"),
	)

	// Ventana de scroll para la lista
	visibleH := height - 2
	start := m.cursor - visibleH + 1
	if start < 0 {
		start = 0
	}
	if m.cursor < start {
		start = m.cursor
	}

	var rows []string
	rows = append(rows, header)
	for i, p := range m.state.Filtered {
		if i < start || i >= start+visibleH {
			continue
		}
		rows = append(rows, m.renderRow(i, p, width-4))
	}

	// Indicador de búsqueda activa
	if m.searching {
		prompt := fmt.Sprintf("Buscar: %s▋", m.state.SearchQuery)
		rows = append([]string{lipgloss.NewStyle().Foreground(lipgloss.Color("212")).Render(prompt)}, rows...)
	}

	return border.Render(lipgloss.JoinVertical(lipgloss.Left, rows...))
}

func (m Model) renderRow(idx int, p domain.Package, width int) string {
	checkbox := "☐"
	if m.state.Selected[idx] {
		checkbox = "☒"
	}
	upgrade := ""
	if p.NewVersion != "" {
		upgrade = " ↑"
	}
	row := fmt.Sprintf(" %s %-28s %-8s %s%s", checkbox, p.Name, p.Manager, p.Version, upgrade)
	if len(row) > width {
		row = row[:width]
	}

	cursor := lipgloss.NewStyle().Background(lipgloss.Color("238"))
	sel := lipgloss.NewStyle().Foreground(lipgloss.Color("212"))

	if idx == m.cursor {
		return cursor.Render(row)
	}
	if m.state.Selected[idx] {
		return sel.Render(row)
	}
	return row
}

func (m Model) renderDetail(width, height int) string {
	active := m.state.ActivePanel == domain.PanelDetail
	style := lipgloss.NewStyle().
		Width(width - 2).
		Height(height).
		Border(lipgloss.RoundedBorder()).
		Padding(0, 1)
	if active {
		style = style.BorderForeground(lipgloss.Color("86"))
	}

	if len(m.state.Filtered) == 0 || m.cursor >= len(m.state.Filtered) {
		return style.Render(lipgloss.NewStyle().Faint(true).Render("Sin paquetes"))
	}

	p := m.state.Filtered[m.cursor]
	title := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("86")).Render(p.Name)

	lines := []string{title, ""}
	if p.NewVersion != "" {
		lines = append(lines,
			fmt.Sprintf("Versión:  %s", p.Version),
			lipgloss.NewStyle().Foreground(lipgloss.Color("226")).Render(
				fmt.Sprintf("Nueva:    %s ↑", p.NewVersion),
			),
		)
	} else {
		lines = append(lines, fmt.Sprintf("Versión:  %s", p.Version))
	}
	lines = append(lines,
		fmt.Sprintf("Gestor:   %s", p.Manager),
		fmt.Sprintf("Origen:   %s", p.Origin),
		fmt.Sprintf("Tamaño:   %s", formatSize(p.Size)),
		fmt.Sprintf("Nativo:   %s", boolStr(p.IsNative)),
		"",
		p.Description,
	)

	return style.Render(lipgloss.JoinVertical(lipgloss.Left, lines...))
}

func (m Model) renderHelpBar() string {
	style := lipgloss.NewStyle().Faint(true)
	help := "[↑↓] Navegar  [Space] Selec  [/] Buscar  [s] Ordenar  [d] Desinstalar  [u] Actualizar  [Tab] Panel  [q] Salir"
	if m.searching {
		help = "[Esc/Enter] Cerrar búsqueda  [Backspace] Borrar"
	}
	return style.Render(help)
}

func formatSize(b int64) string {
	switch {
	case b >= 1_000_000_000:
		return fmt.Sprintf("%.1f GB", float64(b)/1_000_000_000)
	case b >= 1_000_000:
		return fmt.Sprintf("%.1f MB", float64(b)/1_000_000)
	case b >= 1_000:
		return fmt.Sprintf("%.1f KB", float64(b)/1_000)
	default:
		return fmt.Sprintf("%d B", b)
	}
}

func boolStr(b bool) string {
	if b {
		return "Sí"
	}
	return "No"
}
```

- [ ] **Step 4: Verificar que los tests pasan**

```bash
go test ./internal/ui/...
```

Expected: `ok  github.com/hbustos/pkgsh/internal/ui`

- [ ] **Step 5: Commit**

```bash
git add internal/ui/app.go internal/ui/app_test.go
git commit -m "feat(ui): add root Model with full layout and keybindings"
```

---

### Task 6: Actualizar cmd/pkgsh/main.go

**Files:**
- Modify: `cmd/pkgsh/main.go`

- [ ] **Step 1: Reemplazar el contenido de main.go**

```go
package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/hbustos/pkgsh/internal/domain"
	"github.com/hbustos/pkgsh/internal/ui"
)

var version = "dev"

// mockPackages son los paquetes de prueba para PR 1.
// Se reemplazarán por el adapter real en PR 3.
var mockPackages = []domain.Package{
	{Name: "apt",     Version: "2.6.1",   Manager: domain.ManagerApt,      Size: 4_096_000,   Description: "Advanced Package Tool",               IsNative: true,  Origin: "ubuntu"},
	{Name: "bash",    Version: "5.2.15",  Manager: domain.ManagerApt,      Size: 1_800_000,   Description: "GNU Bourne Again SHell",               IsNative: true,  Origin: "ubuntu"},
	{Name: "curl",    Version: "7.88.1",  Manager: domain.ManagerApt,      Size: 448_000,     Description: "Command line tool for transferring data", IsNative: true, Origin: "ubuntu"},
	{Name: "firefox", Version: "126.0.1", NewVersion: "127.0", Manager: domain.ManagerApt, Size: 245_000_000, Description: "Web browser from Mozilla", IsNative: true, Origin: "ubuntu"},
	{Name: "git",     Version: "2.39.2",  Manager: domain.ManagerApt,      Size: 22_000_000,  Description: "Fast, scalable, distributed revision control", IsNative: true, Origin: "ubuntu"},
	{Name: "node",    Version: "20.11.0", Manager: domain.ManagerNpm,      Size: 98_000_000,  Description: "JavaScript runtime (global)",          IsNative: false, Origin: "npmjs"},
	{Name: "python3", Version: "3.11.6",  Manager: domain.ManagerApt,      Size: 6_500_000,   Description: "Interactive high-level object-oriented language", IsNative: true, Origin: "ubuntu"},
	{Name: "snapd",   Version: "2.63",    Manager: domain.ManagerSnap,     Size: 87_000_000,  Description: "Snap daemon and tooling",              IsNative: false, Origin: "snap-store"},
	{Name: "vim",     Version: "2:9.0",   Manager: domain.ManagerApt,      Size: 3_400_000,   Description: "Vi IMproved - enhanced vi editor",     IsNative: true,  Origin: "ubuntu"},
	{Name: "vlc",     Version: "3.0.20",  Manager: domain.ManagerApt,      Size: 45_000_000,  Description: "Multimedia player and framework",      IsNative: true,  Origin: "ubuntu"},
	{Name: "wget",    Version: "1.21.3",  Manager: domain.ManagerApt,      Size: 1_200_000,   Description: "Retrieves files from the web",         IsNative: true,  Origin: "ubuntu"},
	{Name: "zsh",     Version: "5.9",     Manager: domain.ManagerApt,      Size: 2_900_000,   Description: "Shell with lots of features",          IsNative: true,  Origin: "ubuntu"},
}

func main() {
	var (
		managerFlag  = flag.String("manager", "", "filtrar por gestor (comma-separated: apt,snap,...)")
		upgradeable  = flag.Bool("upgradeable", false, "mostrar solo paquetes con actualizaciones")
		native       = flag.Bool("native", false, "mostrar solo paquetes nativos del OS")
		searchFlag   = flag.String("search", "", "arrancar con búsqueda activa")
		ver          = flag.Bool("version", false, "mostrar versión")
	)
	flag.Parse()

	if *ver {
		fmt.Printf("pkgsh %s\n", version)
		os.Exit(0)
	}

	pkgs := mockPackages

	// Aplicar filtros de flags de arranque
	if *upgradeable {
		var filtered []domain.Package
		for _, p := range pkgs {
			if p.NewVersion != "" {
				filtered = append(filtered, p)
			}
		}
		pkgs = filtered
	}
	if *native {
		var filtered []domain.Package
		for _, p := range pkgs {
			if p.IsNative {
				filtered = append(filtered, p)
			}
		}
		pkgs = filtered
	}

	model := ui.New(pkgs)

	// Aplicar flags de arranque al estado inicial
	if *searchFlag != "" {
		// El modelo aplica el filtro internamente al recibir SearchQuery no vacío.
		// Por ahora pasamos directamente los paquetes filtrados.
		filtered := domain.Filter(pkgs, *searchFlag, "")
		model = ui.New(filtered)
	}
	if *managerFlag != "" {
		parts := strings.Split(*managerFlag, ",")
		if len(parts) == 1 {
			filtered := domain.Filter(pkgs, "", domain.ManagerType(parts[0]))
			model = ui.New(filtered)
		}
	}

	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
```

- [ ] **Step 2: Compilar**

```bash
go build ./cmd/pkgsh
```

Expected: sin output (sin errores)

- [ ] **Step 3: Smoke test manual**

```bash
go run ./cmd/pkgsh
```

Expected: el TUI abre en fullscreen mostrando la lista de 12 paquetes mock. Verifica:
- `↑↓` mueve el cursor
- `Space` marca/desmarca con `☒`
- `2` filtra solo paquetes apt
- `1` vuelve a todos
- `/` activa búsqueda, escribir "fir" filtra a firefox
- `Esc` limpia búsqueda
- `d` abre modal de confirmación
- `s` confirma y aparecen líneas en el log panel
- `q` sale

- [ ] **Step 4: Commit**

```bash
git add cmd/pkgsh/main.go
git commit -m "feat(main): launch TUI with mock packages"
```

---

### Task 7: Abrir PR 1

- [ ] **Step 1: Push de la rama**

```bash
git push -u origin feat/tui-mock
```

- [ ] **Step 2: Crear PR**

```bash
gh pr create --base master \
  --title "feat(ui): full Bubble Tea TUI with mock data" \
  --body "$(cat <<'EOF'
## Summary

- Adds full Bubble Tea TUI: tab bar, list panel, detail panel, log panel, modals
- List panel: navigation (↑↓), multi-select (Space/a), search (/), sort (s), tab filters (1-8)
- Detail panel: shows name, version, upgrade indicator, size, origin, description
- Log panel: streams simulated operation output line by line
- Modal: confirmation for remove/update, quit-with-operation confirmation
- 12 hardcoded mock packages to validate all interactions without system calls
- Replaces the `pkgsh: TUI en construcción` stub in main.go

## Test plan

- [ ] `go test ./...` passes
- [ ] `go build ./cmd/pkgsh` compiles clean
- [ ] `go run ./cmd/pkgsh` shows TUI with 12 packages
- [ ] All keybindings in the help bar work as documented

🤖 Generated with [Claude Code](https://claude.com/claude-code)
EOF
)"
```

---

## PR 2 — Adapter apt real

**Branch:** `feat/adapter-apt`  
> Crear desde `master` (no desde `feat/tui-mock` — son independientes).

---

### Task 8: Crear rama y test de parsing con fixture

**Files:**
- Modify: `internal/adapters/apt/adapter.go`
- Create: `internal/adapters/apt/adapter_test.go`

- [ ] **Step 1: Crear rama desde master**

```bash
git checkout master
git checkout -b feat/adapter-apt
```

- [ ] **Step 2: Escribir TestListParsing con fixture**

Crear `internal/adapters/apt/adapter_test.go`:

```go
package apt_test

import (
	"testing"

	"github.com/hbustos/pkgsh/internal/adapters/apt"
	"github.com/hbustos/pkgsh/internal/domain"
)

// dpkgFixture simula la salida de:
// dpkg-query -W -f='${Package}\t${Version}\t${Installed-Size}\t${db:Status-Abbrev}\t${binary:Summary}\n'
const dpkgFixture = `firefox	126.0.1	245000	ii 	Web browser from Mozilla
vlc	3.0.20	45000	ii 	Multimedia player and framework
broken-pkg	1.0	100	rc 	Residual config only, should be skipped
snapd	2.63	87000	ii 	Snap daemon and tooling
`

// upgradeableFixture simula la salida de: apt list --upgradeable
const upgradeableFixture = `Listing... Done
firefox/jammy-updates 127.0 amd64 [upgradable from: 126.0.1]
`

func TestParseInstalled(t *testing.T) {
	pkgs := apt.ParseInstalled([]byte(dpkgFixture))

	if len(pkgs) != 3 {
		t.Fatalf("expected 3 packages (skipping rc status), got %d: %v", len(pkgs), pkgs)
	}

	firefox := pkgs[0]
	if firefox.Name != "firefox" {
		t.Errorf("expected firefox, got %s", firefox.Name)
	}
	if firefox.Version != "126.0.1" {
		t.Errorf("expected version 126.0.1, got %s", firefox.Version)
	}
	if firefox.Size != 245000*1024 {
		t.Errorf("expected size %d, got %d", 245000*1024, firefox.Size)
	}
	if firefox.Manager != domain.ManagerApt {
		t.Errorf("expected manager apt, got %s", firefox.Manager)
	}
	if firefox.Description != "Web browser from Mozilla" {
		t.Errorf("unexpected description: %s", firefox.Description)
	}
}

func TestParseUpgradeable(t *testing.T) {
	newVersions := apt.ParseUpgradeable([]byte(upgradeableFixture))

	if newVersions["firefox"] != "127.0" {
		t.Errorf("expected firefox new version 127.0, got %q", newVersions["firefox"])
	}
	if _, ok := newVersions["vlc"]; ok {
		t.Error("vlc should not be in upgradeable map")
	}
}

func TestListIntegration(t *testing.T) {
	// Este test requiere Linux con dpkg-query disponible.
	if _, err := exec.LookPath("dpkg-query"); err != nil {
		t.Skip("dpkg-query not available")
	}

	a := apt.New()
	pkgs, err := a.List()
	if err != nil {
		t.Fatalf("List() error: %v", err)
	}
	if len(pkgs) == 0 {
		t.Error("expected at least one installed package")
	}
	for _, p := range pkgs {
		if p.Manager != domain.ManagerApt {
			t.Errorf("expected all packages to have manager apt, got %s for %s", p.Manager, p.Name)
		}
		if p.Name == "" {
			t.Error("found package with empty name")
		}
	}
}
```

> **Nota:** El test de integración usa `exec.LookPath` — agrega `"os/exec"` a los imports.

- [ ] **Step 3: Verificar que falla**

```bash
go test ./internal/adapters/apt/...
```

Expected: `FAIL — undefined: apt.ParseInstalled`

---

### Task 9: Implementar parseInstalled y parseUpgradeable

**Files:**
- Modify: `internal/adapters/apt/adapter.go`

- [ ] **Step 1: Reemplazar el contenido de adapter.go**

```go
package apt

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	"github.com/hbustos/pkgsh/internal/domain"
)

type Adapter struct{}

func New() *Adapter { return &Adapter{} }

func (a *Adapter) Name() string { return "apt" }

// List retorna todos los paquetes instalados vía dpkg-query,
// enriquecidos con NewVersion de apt list --upgradeable.
func (a *Adapter) List() ([]domain.Package, error) {
	out, err := exec.Command(
		"dpkg-query", "-W",
		`-f=${Package}\t${Version}\t${Installed-Size}\t${db:Status-Abbrev}\t${binary:Summary}\n`,
	).Output()
	if err != nil {
		return nil, fmt.Errorf("dpkg-query: %w", err)
	}

	pkgs := ParseInstalled(out)

	upgradeOut, _ := exec.Command("apt", "list", "--upgradeable").Output()
	newVersions := ParseUpgradeable(upgradeOut)
	for i := range pkgs {
		if v, ok := newVersions[pkgs[i].Name]; ok {
			pkgs[i].NewVersion = v
		}
	}

	return pkgs, nil
}

// ParseInstalled parsea la salida de dpkg-query en paquetes.
// Exportado para tests.
func ParseInstalled(data []byte) []domain.Package {
	var pkgs []domain.Package
	for _, line := range strings.Split(strings.TrimSpace(string(data)), "\n") {
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "\t", 5)
		if len(parts) < 5 {
			continue
		}
		name, version, sizeStr, statusAbbrev, desc := parts[0], parts[1], parts[2], parts[3], parts[4]
		// Status abbreviation: first char 'i' = installed correctly
		if len(statusAbbrev) == 0 || statusAbbrev[0] != 'i' {
			continue
		}
		sizeKB, _ := strconv.ParseInt(strings.TrimSpace(sizeStr), 10, 64)
		pkgs = append(pkgs, domain.Package{
			Name:        name,
			Version:     version,
			Manager:     domain.ManagerApt,
			Size:        sizeKB * 1024,
			Description: desc,
			IsNative:    true,
		})
	}
	return pkgs
}

// ParseUpgradeable parsea la salida de apt list --upgradeable.
// Retorna un mapa de nombre de paquete → nueva versión.
// Exportado para tests.
func ParseUpgradeable(data []byte) map[string]string {
	result := make(map[string]string)
	for i, line := range strings.Split(strings.TrimSpace(string(data)), "\n") {
		if i == 0 || line == "" {
			continue // saltar "Listing... Done"
		}
		// formato: "name/repo version arch [upgradable from: oldver]"
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		name := strings.SplitN(fields[0], "/", 2)[0]
		newVersion := fields[1]
		result[name] = newVersion
	}
	return result
}

// Remove desinstala los paquetes dados vía sudo apt remove.
// El output se streams al Operation retornado.
func (a *Adapter) Remove(pkgs []domain.Package) *domain.Operation {
	op := domain.NewOperation()
	names := packageNames(pkgs)
	args := append([]string{"apt", "remove", "-y"}, names...)
	go runCommand(op, "sudo", args...)
	return op
}

// Update actualiza los paquetes dados vía sudo apt-get install --only-upgrade.
func (a *Adapter) Update(pkgs []domain.Package) *domain.Operation {
	op := domain.NewOperation()
	names := packageNames(pkgs)
	args := append([]string{"apt-get", "install", "--only-upgrade", "-y"}, names...)
	go runCommand(op, "sudo", args...)
	return op
}

// Info retorna información detallada de un paquete vía apt-cache show.
func (a *Adapter) Info(pkg domain.Package) (domain.PackageInfo, error) {
	out, err := exec.Command("apt-cache", "show", pkg.Name).Output()
	if err != nil {
		return domain.PackageInfo{}, fmt.Errorf("apt-cache show: %w", err)
	}
	info := domain.PackageInfo{Package: pkg}
	for _, line := range strings.Split(string(out), "\n") {
		if strings.HasPrefix(line, "Depends: ") {
			deps := strings.TrimPrefix(line, "Depends: ")
			for _, dep := range strings.Split(deps, ", ") {
				info.Depends = append(info.Depends, strings.Fields(dep)[0])
			}
		}
	}
	return info, nil
}

// runCommand ejecuta el comando y copia stdout+stderr al Operation.
func runCommand(op *domain.Operation, name string, args ...string) {
	cmd := exec.Command(name, args...)
	cmd.Stdout = op.Writer()
	cmd.Stderr = op.Writer()
	err := cmd.Run()
	op.Done(err)
}

func packageNames(pkgs []domain.Package) []string {
	names := make([]string, len(pkgs))
	for i, p := range pkgs {
		names[i] = p.Name
	}
	return names
}
```

- [ ] **Step 2: Corregir el import faltante en adapter_test.go**

El test de integración usa `exec.LookPath`. Agrega `"os/exec"` a los imports de `adapter_test.go`:

```go
import (
	"os/exec"
	"testing"

	"github.com/hbustos/pkgsh/internal/adapters/apt"
	"github.com/hbustos/pkgsh/internal/domain"
)
```

- [ ] **Step 3: Correr los tests**

```bash
go test ./internal/adapters/apt/...
```

Expected:
```
ok  github.com/hbustos/pkgsh/internal/adapters/apt
```

(El test de integración corre si `dpkg-query` está disponible, o se salta con `SKIP` si no lo está.)

- [ ] **Step 4: Verificar que el proyecto compila**

```bash
go build ./...
```

Expected: sin output

- [ ] **Step 5: Commit**

```bash
git add internal/adapters/apt/adapter.go internal/adapters/apt/adapter_test.go
git commit -m "feat(adapter/apt): implement List, Remove, Update, Info"
```

---

### Task 10: Abrir PR 2

- [ ] **Step 1: Push de la rama**

```bash
git push -u origin feat/adapter-apt
```

- [ ] **Step 2: Crear PR**

```bash
gh pr create --base master \
  --title "feat(adapter/apt): implement apt adapter with tests" \
  --body "$(cat <<'EOF'
## Summary

- Implements `List()` using `dpkg-query` + `apt list --upgradeable`
- Exports `ParseInstalled()` and `ParseUpgradeable()` for unit testing
- Unit tests with hardcoded fixtures (no system calls required)
- Integration test with `t.Skip` if `dpkg-query` not available
- `Remove()` and `Update()` stream stdout/stderr to `*domain.Operation`
- `Info()` parses `apt-cache show` for dependency info
- Security invariant: all commands built as `[]string`, never shell interpolation

## Test plan

- [ ] `go test ./internal/adapters/apt/...` passes (unit tests always, integration if dpkg available)
- [ ] `go build ./...` compiles clean

🤖 Generated with [Claude Code](https://claude.com/claude-code)
EOF
)"
```

---

## PR 3 — Conectar adapter apt con TUI

**Branch:** `feat/connect-apt-tui`  
> Crear desde `master` **después de que PR 1 y PR 2 estén mergeados**.

---

### Task 11: Agregar carga progresiva a app.go

**Files:**
- Modify: `internal/ui/app.go`

- [ ] **Step 1: Agregar mensajes y estado de carga**

Agregar al bloque de tipos al inicio de `app.go` (después de `tickMsg`):

```go
// packagesLoadedMsg llega cuando un adapter termina de cargar sus paquetes.
type packagesLoadedMsg struct {
	pkgs []domain.Package
	err  error
}

// loadingMsg se usa para mostrar el spinner mientras cargan los adapters.
type loadingMsg struct{}
```

- [ ] **Step 2: Agregar campo loading al Model**

En la struct `Model`, agregar `loading bool`:

```go
type Model struct {
	state     domain.AppState
	cursor    int
	width     int
	height    int
	log       LogModel
	modal     ModalModel
	showModal bool
	searching bool
	loading   bool  // true mientras los adapters cargan
}
```

- [ ] **Step 3: Agregar función New con loading inicial**

Agregar función `NewLoading()` para inicializar el modelo en estado de carga:

```go
// NewLoading crea un modelo vacío con el spinner de carga activo.
// Se usa cuando los paquetes aún no están disponibles (PR 3).
func NewLoading() Model {
	state := domain.AppState{
		Packages: nil,
		Filtered: nil,
		Selected: make(map[int]bool),
		SortBy:   domain.SortByName,
	}
	return Model{
		state:   state,
		log:     newLogModel(),
		width:   80,
		height:  24,
		loading: true,
	}
}
```

- [ ] **Step 4: Manejar packagesLoadedMsg en Update**

En el `switch msg.(type)` de `Update`, agregar antes del case `tea.KeyMsg`:

```go
case packagesLoadedMsg:
	if msg.err != nil {
		m.log = m.log.appendLine("Error cargando paquetes: " + msg.err.Error())
	} else {
		m.state.Packages = append(m.state.Packages, msg.pkgs...)
		m.applyFilter()
		if m.cursor >= len(m.state.Filtered) {
			m.cursor = 0
		}
	}
	m.loading = len(m.state.Packages) == 0
```

- [ ] **Step 5: Mostrar spinner en View cuando loading=true**

Al inicio del método `View()`, reemplazar:

```go
if m.width == 0 {
    return "Cargando..."
}
```

por:

```go
if m.width == 0 {
    return "Cargando..."
}
if m.loading && len(m.state.Packages) == 0 {
    return lipgloss.NewStyle().
        Width(m.width).Height(m.height).
        Align(lipgloss.Center, lipgloss.Center).
        Render("Cargando paquetes...")
}
```

- [ ] **Step 6: Compilar y pasar tests**

```bash
go build ./...
go test ./internal/ui/...
```

Expected: sin errores

- [ ] **Step 7: Commit**

```bash
git add internal/ui/app.go
git commit -m "feat(ui): add progressive loading support for real adapters"
```

---

### Task 12: Actualizar main.go con adapter real

**Files:**
- Modify: `cmd/pkgsh/main.go`

- [ ] **Step 1: Reemplazar main.go**

```go
package main

import (
	"flag"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/hbustos/pkgsh/internal/adapters/apt"
	"github.com/hbustos/pkgsh/internal/domain"
	"github.com/hbustos/pkgsh/internal/ui"
)

var version = "dev"

func main() {
	var (
		// Flags reservados para aplicar en un PR futuro una vez que
		// el modelo soporte recibirlos como estado inicial.
		_ = flag.String("manager", "", "filtrar por gestor (comma-separated: apt,snap,...)")
		_ = flag.Bool("upgradeable", false, "mostrar solo paquetes con actualizaciones")
		_ = flag.Bool("native", false, "mostrar solo paquetes nativos del OS")
		_ = flag.String("search", "", "arrancar con búsqueda activa")
		ver = flag.Bool("version", false, "mostrar versión")
	)
	flag.Parse()

	if *ver {
		fmt.Printf("pkgsh %s\n", version)
		os.Exit(0)
	}

	adapters := []domain.PackageManager{
		apt.New(),
	}

	model := ui.NewLoading()

	p := tea.NewProgram(model, tea.WithAltScreen())

	// Cargar paquetes en goroutine y enviar al programa cuando terminen.
	go func() {
		msg := loadPackages(adapters)()
		p.Send(msg)
	}()

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

// loadPackages retorna un tea.Cmd que carga paquetes de todos los adapters
// y envía un packagesLoadedMsg por adapter.
func loadPackages(adapters []domain.PackageManager) tea.Cmd {
	return func() tea.Msg {
		// Por ahora un solo adapter; en el futuro usar goroutines paralelas
		// y enviar múltiples mensajes.
		a := adapters[0]
		pkgs, err := a.List()
		return ui.PackagesLoadedMsg{Pkgs: pkgs, Err: err}
	}
}
```

> **Nota:** `ui.PackagesLoadedMsg` necesita ser exportado desde `app.go`. Ver paso siguiente.

- [ ] **Step 2: Exportar PackagesLoadedMsg en app.go**

En `internal/ui/app.go`, cambiar `packagesLoadedMsg` por `PackagesLoadedMsg` (mayúscula) para que `main.go` pueda construirlo:

```go
// PackagesLoadedMsg llega cuando un adapter termina de cargar sus paquetes.
type PackagesLoadedMsg struct {
	Pkgs []domain.Package
	Err  error
}
```

Y en el `Update` de `app.go`, actualizar el case:

```go
case PackagesLoadedMsg:
	if msg.Err != nil {
		m.log = m.log.appendLine("Error cargando paquetes: " + msg.Err.Error())
	} else {
		m.state.Packages = append(m.state.Packages, msg.Pkgs...)
		m.applyFilter()
		if m.cursor >= len(m.state.Filtered) {
			m.cursor = 0
		}
	}
	m.loading = len(m.state.Packages) == 0
```

- [ ] **Step 3: Actualizar tests de app_test.go si es necesario**

Los tests existentes usan `New(pkgs)` que no cambia — solo se agregó `NewLoading()`. Los tests no deben romperse. Verificar:

```bash
go test ./...
```

Expected: `ok` en todos los packages

- [ ] **Step 4: Smoke test final**

```bash
go run ./cmd/pkgsh
```

Expected: el TUI abre, muestra "Cargando paquetes..." brevemente, luego se puebla con los paquetes reales instalados en el sistema via `dpkg-query`. Verificar que:
- Hay paquetes reales (apt, bash, curl, etc.) en la lista
- `firefox` tiene `↑` si hay actualización disponible
- Todas las interacciones del PR 1 siguen funcionando

- [ ] **Step 5: Commit**

```bash
git add cmd/pkgsh/main.go internal/ui/app.go
git commit -m "feat(main): wire apt adapter with real package loading"
```

---

### Task 13: Abrir PR 3

- [ ] **Step 1: Crear rama y push**

```bash
git checkout -b feat/connect-apt-tui
git push -u origin feat/connect-apt-tui
```

- [ ] **Step 2: Crear PR**

```bash
gh pr create --base master \
  --title "feat: connect apt adapter to TUI (working vertical slice)" \
  --body "$(cat <<'EOF'
## Summary

- Replaces mock package data with real `apt.Adapter.List()` call
- Adds `NewLoading()` model + `PackagesLoadedMsg` for progressive loading
- TUI shows "Cargando paquetes..." spinner while `dpkg-query` runs
- Packages populate progressively as adapters finish (architecture ready for multiple adapters)
- All keybindings, search, filter, sort, and simulated operations from PR 1 work unchanged

**After this PR:** `pkgsh` is a working tool that shows real installed apt packages with full TUI interaction.

## Test plan

- [ ] `go test ./...` passes
- [ ] `go run ./cmd/pkgsh` shows real system packages
- [ ] Remove operation streams output to log panel
- [ ] Search and filter work on real data

🤖 Generated with [Claude Code](https://claude.com/claude-code)
EOF
)"
```

---

## Notas finales

### Sobre sudo en operaciones Remove/Update

Las operaciones de PR 2/3 corren `sudo apt remove/upgrade` directamente. Esto funciona si:
- El usuario tiene credenciales de sudo cacheadas, o
- El usuario tiene `NOPASSWD` configurado para apt en `/etc/sudoers`

El modal de sudo password (ya implementado en `modal.go`) se conectará al stdin del proceso en un PR futuro mediante `cmd.Stdin = stdinPipe` donde `stdinPipe` recibe la contraseña del modal.

### Orden de merge

```
PR 1 (feat/tui-mock) → merge a master
PR 2 (feat/adapter-apt) → merge a master
PR 3 (feat/connect-apt-tui) → crear DESPUÉS de que PR1 y PR2 estén en master
```
