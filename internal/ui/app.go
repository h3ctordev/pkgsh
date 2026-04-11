package ui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/hbustos/pkgsh/internal/domain"
)

type Options struct {
	Manager      domain.ManagerType
	Upgradeable  bool
	Native       bool
	Search       string
	SecurityMode bool
}

type opKind int

const (
	opRemove opKind = iota
	opUpdate
)

type pendingOp struct {
	manager domain.ManagerType
	pkgs    []domain.Package
}

type spinnerTickMsg struct{}
type logCollapseMsg struct{}

type AppModel struct {
	state          domain.AppState
	list           ListModel
	detail         DetailModel
	log            LogModel
	modal          *ModalModel
	searching      bool
	width          int
	height         int
	adapters       map[domain.ManagerType]domain.PackageManager
	pendingOps     []pendingOp
	currentKind    opKind
	currentManager domain.ManagerType
	logCollapsed   bool
	spinnerFrame   int
}

func New(pkgs []domain.Package, adapters map[domain.ManagerType]domain.PackageManager, opts Options) AppModel {
	state := domain.AppState{
		Packages:     pkgs,
		Selected:     make(map[int]bool),
		SortBy:       domain.SortByName,
		ActivePanel:  domain.PanelList,
		SearchQuery:  opts.Search,
		ActiveTab:    opts.Manager,
		SecurityMode: opts.SecurityMode,
	}

	filtered := domain.Filter(pkgs, state.SearchQuery, state.ActiveTab, state.SecurityMode)
	if opts.Upgradeable {
		filtered = filterUpgradeable(filtered)
	}
	if opts.Native {
		filtered = filterNative(filtered)
	}
	filtered = domain.Sort(filtered, state.SortBy)
	state.Filtered = filtered

	return AppModel{
		state:        state,
		list:         newListModel().SetItems(filtered),
		detail:       newDetailModel(),
		log:          newLogModel(),
		adapters:     adapters,
		logCollapsed: true,
	}
}

func filterUpgradeable(pkgs []domain.Package) []domain.Package {
	var out []domain.Package
	for _, p := range pkgs {
		if p.NewVersion != "" {
			out = append(out, p)
		}
	}
	return out
}

func filterNative(pkgs []domain.Package) []domain.Package {
	var out []domain.Package
	for _, p := range pkgs {
		if p.IsNative {
			out = append(out, p)
		}
	}
	return out
}

func (m AppModel) Init() tea.Cmd { return nil }

func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.modal != nil {
		return m.updateModal(msg)
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case spinnerTickMsg:
		if m.state.Operation != nil {
			m.spinnerFrame = (m.spinnerFrame + 1) % 8
			return m, spinnerTickCmd()
		}
		return m, nil

	case logCollapseMsg:
		if m.state.Operation == nil {
			m.logCollapsed = true
		}
		return m, nil

	case operationLineMsg:
		line := string(msg)
		if strings.Contains(line, "PKGSH_SUDO:") && m.state.Operation != nil {
			modal := newSudoModal()
			m.modal = &modal
			return m, readLineCmd(m.state.Operation)
		}
		m.log = m.log.appendLine(line)
		return m, readLineCmd(m.state.Operation)

	case operationDoneMsg:
		if msg.err != nil {
			m.log = m.log.appendLine(fmt.Sprintf("[ERROR] %s: %v", m.currentManager, msg.err))
		} else {
			m.log = m.log.appendLine("✓ Listo.")
		}
		m.state.Operation = nil
		var cmd tea.Cmd
		m, cmd = m.startNextOp()
		if m.state.Operation == nil {
			return m, tea.Batch(cmd, logCollapseCmd())
		}
		return m, cmd

	case packagesReloadedMsg:
		m.state.Packages = msg.pkgs
		m = m.applyFilter()
		return m, nil

	case tea.KeyMsg:
		if m.searching {
			return m.updateSearch(msg)
		}
		return m.updateKeys(msg)
	}

	return m, nil
}

func spinnerTickCmd() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(_ time.Time) tea.Msg {
		return spinnerTickMsg{}
	})
}

func logCollapseCmd() tea.Cmd {
	return func() tea.Msg {
		time.Sleep(3 * time.Second)
		return logCollapseMsg{}
	}
}

func (m AppModel) updateModal(msg tea.Msg) (tea.Model, tea.Cmd) {
	updated, confirmed, cancelled := m.modal.Update(msg)
	if cancelled {
		m.modal = nil
		return m, nil
	}
	if confirmed {
		switch updated.modalType {
		case ModalConfirm:
			sel := m.list.AllSelected(m.state.Packages)
			m.list = m.list.ClearSelection()
			m.modal = nil

			if m.state.SecurityMode {
				var allowed []domain.Package
				for _, p := range sel {
					if domain.IsSystemPackage(p) {
						m.log = m.log.appendLine(fmt.Sprintf("[SECURITY] %s: paquete del sistema, operación bloqueada", p.Name))
					} else {
						allowed = append(allowed, p)
					}
				}
				sel = allowed
			}
			if len(sel) == 0 {
				return m, nil
			}

			order := []domain.ManagerType{
				domain.ManagerApt, domain.ManagerSnap, domain.ManagerFlatpak,
				domain.ManagerDpkg, domain.ManagerPip, domain.ManagerNpm, domain.ManagerAppImage,
			}
			byManager := make(map[domain.ManagerType][]domain.Package)
			for _, p := range sel {
				byManager[p.Manager] = append(byManager[p.Manager], p)
			}
			for _, mgr := range order {
				if pkgs, ok := byManager[mgr]; ok {
					m.pendingOps = append(m.pendingOps, pendingOp{manager: mgr, pkgs: pkgs})
				}
			}

			var cmd tea.Cmd
			m, cmd = m.startNextOp()
			return m, cmd

		case ModalSudo:
			m.modal = nil
			return m, nil

		case ModalQuitConfirm:
			return m, tea.Quit
		}
		return m, nil
	}
	m.modal = &updated
	return m, nil
}

func (m AppModel) updateSearch(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEnter, tea.KeyEsc:
		m.searching = false
	case tea.KeyBackspace:
		if len(m.state.SearchQuery) > 0 {
			m.state.SearchQuery = m.state.SearchQuery[:len(m.state.SearchQuery)-1]
			m = m.applyFilter()
		}
	default:
		if msg.Type == tea.KeyRunes {
			m.state.SearchQuery += string(msg.Runes)
			m = m.applyFilter()
		}
	}
	return m, nil
}

func (m AppModel) updateKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyCtrlC:
		return m, tea.Quit

	case tea.KeyTab:
		m.state.ActivePanel = (m.state.ActivePanel + 1) % 3
		return m, nil

	case tea.KeyUp, tea.KeyDown:
		switch m.state.ActivePanel {
		case domain.PanelList:
			var cmd tea.Cmd
			m.list, cmd = m.list.Update(msg)
			return m, cmd
		case domain.PanelLog:
			var cmd tea.Cmd
			m.log, cmd = m.log.Update(msg)
			return m, cmd
		}

	case tea.KeyPgUp, tea.KeyPgDown:
		if m.state.ActivePanel == domain.PanelLog {
			var cmd tea.Cmd
			m.log, cmd = m.log.Update(msg)
			return m, cmd
		}

	case tea.KeySpace:
		if m.state.ActivePanel == domain.PanelList {
			m.list = m.list.ToggleSelected()
		}

	case tea.KeyEsc:
		m.list = m.list.ClearSelection()

	case tea.KeyRunes:
		switch string(msg.Runes) {
		case "j":
			if m.state.ActivePanel == domain.PanelList {
				var cmd tea.Cmd
				m.list, cmd = m.list.Update(tea.KeyMsg{Type: tea.KeyDown})
				return m, cmd
			}
		case "k":
			if m.state.ActivePanel == domain.PanelList {
				var cmd tea.Cmd
				m.list, cmd = m.list.Update(tea.KeyMsg{Type: tea.KeyUp})
				return m, cmd
			}
		case "g":
			if m.state.ActivePanel == domain.PanelList {
				m.list = m.list.JumpToFirst()
			}
		case "G":
			if m.state.ActivePanel == domain.PanelList {
				m.list = m.list.JumpToLast()
			}
		case "q":
			if m.state.Operation != nil {
				modal := newQuitConfirmModal()
				m.modal = &modal
				return m, nil
			}
			return m, tea.Quit
		case "?":
			modal := newHelpModal()
			m.modal = &modal
			return m, nil
		case "/":
			m.searching = true
		case "a":
			m.list = m.list.SelectAll()
		case "r":
			if m.state.Operation == nil {
				return m, m.reloadPackagesCmd()
			}
		case "s":
			m.state.SortBy = (m.state.SortBy + 1) % 4
			m = m.applyFilter()
		case "S":
			m.state.SecurityMode = !m.state.SecurityMode
			m = m.applyFilter()
		case "d":
			sel := m.list.AllSelected(m.state.Packages)
			if len(sel) == 0 {
				return m, nil
			}
			m.currentKind = opRemove
			modal := newConfirmModal(
				fmt.Sprintf("Desinstalar %d paquete(s)", len(sel)),
				strings.Join(packageNames(sel), ", "),
			)
			m.modal = &modal
		case "u":
			sel := m.list.AllSelected(m.state.Packages)
			if len(sel) == 0 {
				return m, nil
			}
			m.currentKind = opUpdate
			modal := newConfirmModal(
				fmt.Sprintf("Actualizar %d paquete(s)", len(sel)),
				strings.Join(packageNames(sel), ", "),
			)
			m.modal = &modal
		case "1":
			m.state.ActiveTab = ""
			m = m.applyFilter()
		case "2":
			m.state.ActiveTab = domain.ManagerApt
			m = m.applyFilter()
		case "3":
			m.state.ActiveTab = domain.ManagerSnap
			m = m.applyFilter()
		case "4":
			m.state.ActiveTab = domain.ManagerFlatpak
			m = m.applyFilter()
		case "5":
			m.state.ActiveTab = domain.ManagerDpkg
			m = m.applyFilter()
		case "6":
			m.state.ActiveTab = domain.ManagerPip
			m = m.applyFilter()
		case "7":
			m.state.ActiveTab = domain.ManagerNpm
			m = m.applyFilter()
		case "8":
			m.state.ActiveTab = domain.ManagerAppImage
			m = m.applyFilter()
		}
	}

	return m, nil
}

func (m AppModel) applyFilter() AppModel {
	filtered := domain.Filter(m.state.Packages, m.state.SearchQuery, m.state.ActiveTab, m.state.SecurityMode)
	filtered = domain.Sort(filtered, m.state.SortBy)
	m.state.Filtered = filtered
	m.list = m.list.SetItems(filtered)
	return m
}

func (m AppModel) startNextOp() (AppModel, tea.Cmd) {
	for len(m.pendingOps) > 0 {
		op := m.pendingOps[0]
		m.pendingOps = m.pendingOps[1:]

		adapter, ok := m.adapters[op.manager]
		if !ok {
			m.log = m.log.appendLine(fmt.Sprintf("[SKIP] %s: adapter no disponible", op.manager))
			continue
		}

		m.currentManager = op.manager
		m.logCollapsed = false
		var operation *domain.Operation
		if m.currentKind == opRemove {
			operation = adapter.Remove(op.pkgs)
		} else {
			operation = adapter.Update(op.pkgs)
		}
		m.state.Operation = operation
		return m, tea.Batch(readLineCmd(operation), spinnerTickCmd())
	}

	m.log = m.log.appendLine("Listo.")
	return m, m.reloadPackagesCmd()
}

type packagesReloadedMsg struct{ pkgs []domain.Package }

func (m AppModel) reloadPackagesCmd() tea.Cmd {
	adapters := m.adapters
	return func() tea.Msg {
		var pkgs []domain.Package
		for _, adapter := range adapters {
			list, _ := adapter.List()
			pkgs = append(pkgs, list...)
		}
		return packagesReloadedMsg{pkgs: pkgs}
	}
}

func packageNames(pkgs []domain.Package) []string {
	names := make([]string, len(pkgs))
	for i, p := range pkgs {
		names[i] = p.Name
	}
	return names
}

func (m AppModel) View() string {
	if m.width == 0 {
		return "Iniciando pkgsh..."
	}

	available := m.height - 8
	if available < 10 {
		available = 10
	}

	var topHeight, logHeight int
	if m.logCollapsed {
		logHeight = 1
		topHeight = available - logHeight
	} else {
		topHeight = available * 3 / 5
		logHeight = available - topHeight
		if logHeight < 3 {
			logHeight = 3
			topHeight = available - logHeight
		}
	}

	listWidth := m.width * 2 / 3
	detailWidth := m.width - listWidth

	sel := m.list.AllSelected(m.state.Packages)
	header := m.viewHeader(len(sel))
	searchBar := m.viewSearchBar(listWidth)
	listView := m.list.View(listWidth, topHeight, m.state.ActivePanel == domain.PanelList)
	detailView := m.detail.View(m.list.CurrentPackage(), detailWidth, topHeight+1, m.state.ActivePanel == domain.PanelDetail)

	body := lipgloss.JoinHorizontal(lipgloss.Top,
		lipgloss.JoinVertical(lipgloss.Left, searchBar, listView),
		detailView,
	)

	selectionBar := m.viewSelectionBar(m.width, sel)
	logView := m.log.View(m.width, logHeight, m.state.ActivePanel == domain.PanelLog, m.logTitle())
	footer := m.viewFooter(len(sel))

	full := lipgloss.JoinVertical(lipgloss.Left, header, body, selectionBar, logView, footer)

	if m.modal != nil {
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center,
			m.modal.View(m.width),
			lipgloss.WithWhitespaceChars(" "),
		)
	}

	return full
}

func (m AppModel) logTitle() string {
	if m.state.Operation == nil {
		return ""
	}
	frames := []string{"⣾", "⣽", "⣻", "⢿", "⡿", "⣟", "⣯", "⣷"}
	frame := frames[m.spinnerFrame%len(frames)]
	cmdName := "remove"
	if m.currentKind == opUpdate {
		cmdName = "update"
	}
	return fmt.Sprintf("%s %s %s", frame, m.currentManager, cmdName)
}

func (m AppModel) viewHeader(selCount int) string {
	tabs := []struct {
		key     string
		label   string
		manager domain.ManagerType
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

	var parts []string
	for _, t := range tabs {
		label := fmt.Sprintf("[%s] %s", t.key, t.label)
		if t.manager == m.state.ActiveTab {
			label = lipgloss.NewStyle().Bold(true).Foreground(colorPrimary).Render(label)
		} else {
			label = lipgloss.NewStyle().Foreground(colorMuted).Render(label)
		}
		parts = append(parts, label)
	}

	upgCount := 0
	for _, p := range m.state.Filtered {
		if p.NewVersion != "" {
			upgCount++
		}
	}
	stats := fmt.Sprintf("%d pkgs", len(m.state.Filtered))
	if upgCount > 0 {
		stats += fmt.Sprintf(" · ↑ %d", upgCount)
	}
	if selCount > 0 {
		stats += fmt.Sprintf(" · sel %d", selCount)
	}
	statsStyled := lipgloss.NewStyle().Foreground(colorMuted).Render(stats)

	title := lipgloss.NewStyle().Bold(true).Foreground(colorPrimary).Render("pkgsh")
	tabsStr := strings.Join(parts, "  ")
	left := title + "  " + tabsStr

	statsWidth := lipgloss.Width(statsStyled)
	leftWidth := lipgloss.Width(left)
	padding := m.width - leftWidth - statsWidth - 2
	if padding < 1 {
		padding = 1
	}

	return lipgloss.NewStyle().
		Width(m.width).
		Background(colorHeaderBg).
		Padding(0, 1).
		Render(left + strings.Repeat(" ", padding) + statsStyled)
}

func (m AppModel) viewSearchBar(width int) string {
	query := m.state.SearchQuery
	if m.searching {
		query += "█"
	}
	return lipgloss.NewStyle().
		Width(width - 2).
		Foreground(colorMuted).
		Padding(0, 1).
		Render("> " + query)
}

func (m AppModel) viewSelectionBar(width int, sel []domain.Package) string {
	base := lipgloss.NewStyle().Width(width).Padding(0, 1)

	if len(sel) == 0 {
		return base.Foreground(colorMuted).Faint(true).Render(
			"Sin selección  ·  Space para seleccionar  ·  a para todo",
		)
	}

	badge := lipgloss.NewStyle().Bold(true).Foreground(colorGreen).Render(fmt.Sprintf("☒ %d", len(sel)))

	var names []string
	sysPkgs := 0
	for _, p := range sel {
		tag := pkgTags(p)
		entry := fmt.Sprintf("%s (%s)", p.Name, string(p.Manager))
		if tag != "" {
			entry += " " + tag
		}
		names = append(names, entry)
		if domain.IsSystemPackage(p) {
			sysPkgs++
		}
	}
	text := truncate(strings.Join(names, "  ·  "), width-12)

	result := base.Render(badge + "  " + text)
	if sysPkgs > 0 {
		warning := lipgloss.NewStyle().Width(width).Padding(0, 1).Foreground(colorSec).Render(
			fmt.Sprintf("⚠  %d paquete(s) del sistema — se bloquearán en modo SEC", sysPkgs),
		)
		result += "\n" + warning
	}
	return result
}

func (m AppModel) viewFooter(selCount int) string {
	var hints string
	if m.state.Operation != nil {
		hints = "[?] Ayuda  [PgUp/PgDn] Scroll log  [q] Confirmar salida"
	} else if selCount > 0 {
		actions := lipgloss.NewStyle().Foreground(colorGreen).Bold(true).Render("[d] Desinstalar  [u] Actualizar")
		hints = "[?] Ayuda  [/] Buscar  " + actions + "  [Esc] Limpiar"
	} else {
		hints = "[?] Ayuda  [/] Buscar  [1-8] Filtro"
	}

	if m.state.SecurityMode {
		sec := lipgloss.NewStyle().Bold(true).Foreground(colorSec).Render("[SEC] ●")
		hints += "  " + sec
	}

	return lipgloss.NewStyle().
		Width(m.width).
		Foreground(colorMuted).
		Faint(true).
		Padding(0, 1).
		Render(hints)
}
