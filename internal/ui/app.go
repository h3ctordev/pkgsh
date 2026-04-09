package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/hbustos/pkgsh/internal/domain"
)

// Options configuran el arranque de la aplicación (desde flags CLI).
type Options struct {
	Manager     domain.ManagerType
	Upgradeable bool
	Native      bool
	Search      string
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
	loading        bool
	loadedCount    int
	totalAdapters  int
}

type packagesLoadedMsg struct {
	manager domain.ManagerType
	pkgs    []domain.Package
}

func loadAdapterCmd(mgr domain.ManagerType, adapter domain.PackageManager) tea.Cmd {
	return func() tea.Msg {
		pkgs, _ := adapter.List()
		return packagesLoadedMsg{manager: mgr, pkgs: pkgs}
	}
}

func New(pkgs []domain.Package, adapters map[domain.ManagerType]domain.PackageManager, opts Options) AppModel {
	state := domain.AppState{
		Packages:    pkgs,
		Selected:    make(map[int]bool),
		SortBy:      domain.SortByName,
		ActivePanel: domain.PanelList,
		SearchQuery: opts.Search,
		ActiveTab:   opts.Manager,
	}

	filtered := domain.Filter(pkgs, state.SearchQuery, state.ActiveTab)
	if opts.Upgradeable {
		filtered = filterUpgradeable(filtered)
	}
	if opts.Native {
		filtered = filterNative(filtered)
	}
	filtered = domain.Sort(filtered, state.SortBy)
	state.Filtered = filtered

	m := AppModel{
		state:    state,
		list:     newListModel().SetItems(filtered),
		detail:   newDetailModel(),
		log:      newLogModel(),
		adapters: adapters,
	}

	if pkgs == nil && len(adapters) > 0 {
		m.loading = true
		m.totalAdapters = len(adapters)
	}

	return m
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

func (m AppModel) Init() tea.Cmd {
	if !m.loading {
		return nil
	}
	cmds := make([]tea.Cmd, 0, len(m.adapters))
	for mgr, adapter := range m.adapters {
		cmds = append(cmds, loadAdapterCmd(mgr, adapter))
	}
	return tea.Batch(cmds...)
}

func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.modal != nil {
		return m.updateModal(msg)
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case operationLineMsg:
		m.log = m.log.appendLine(string(msg))
		return m, readLineCmd(m.state.Operation)

	case operationDoneMsg:
		if msg.err != nil {
			m.log = m.log.appendLine(fmt.Sprintf("[ERROR] %s: %v", m.currentManager, msg.err))
		}
		m.state.Operation = nil
		var cmd tea.Cmd
		m, cmd = m.startNextOp()
		return m, cmd

	case packagesLoadedMsg:
		m.state.Packages = append(m.state.Packages, msg.pkgs...)
		m.loadedCount++
		if m.loadedCount >= m.totalAdapters {
			m.loading = false
			m = m.applyFilter()
		}
		return m, nil

	case tea.KeyMsg:
		if m.searching {
			return m.updateSearch(msg)
		}
		return m.updateKeys(msg)
	}

	return m, nil
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

			// Agrupar por manager en orden canónico
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
		case "q":
			if m.state.Operation != nil {
				modal := newQuitConfirmModal()
				m.modal = &modal
				return m, nil
			}
			return m, tea.Quit

		case "/":
			m.searching = true

		case "a":
			m.list = m.list.SelectAll()

		case "s":
			m.state.SortBy = (m.state.SortBy + 1) % 4
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
	filtered := domain.Filter(m.state.Packages, m.state.SearchQuery, m.state.ActiveTab)
	filtered = domain.Sort(filtered, m.state.SortBy)
	m.state.Filtered = filtered
	m.list = m.list.SetItems(filtered)
	return m
}

// startNextOp toma la primera op de pendingOps, arranca la operación y devuelve readLineCmd.
// Si la cola está vacía, loguea "Listo." y devuelve nil.
// Si el adapter del manager no existe, loguea [SKIP] y pasa a la siguiente.
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
		var operation *domain.Operation
		if m.currentKind == opRemove {
			operation = adapter.Remove(op.pkgs)
		} else {
			operation = adapter.Update(op.pkgs)
		}
		m.state.Operation = operation
		return m, readLineCmd(operation)
	}

	m.log = m.log.appendLine("Listo.")
	return m, nil
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

	// Filas consumidas por elementos fuera de los paneles:
	// header(1) + searchBar(1) + selectionBar(1) + bordes lista(2) + bordes log(2) + footer(1) = 8
	available := m.height - 8
	if available < 10 {
		available = 10
	}
	topHeight := available * 3 / 5 // 60% para lista+detalle
	logHeight := available - topHeight
	if logHeight < 3 {
		logHeight = 3
		topHeight = available - logHeight
	}
	listWidth := m.width * 2 / 3
	detailWidth := m.width - listWidth

	header := m.viewHeader()
	searchBar := m.viewSearchBar(listWidth)
	// El panel detalle no tiene searchBar encima, así que tiene 1 fila extra de contenido
	listView := m.list.View(listWidth, topHeight, m.state.ActivePanel == domain.PanelList)
	detailView := m.detail.View(m.list.CurrentPackage(), detailWidth, topHeight+1, m.state.ActivePanel == domain.PanelDetail)

	body := lipgloss.JoinHorizontal(lipgloss.Top,
		lipgloss.JoinVertical(lipgloss.Left, searchBar, listView),
		detailView,
	)

	selectionBar := m.viewSelectionBar(m.width)
	logView := m.log.View(m.width, logHeight, m.state.ActivePanel == domain.PanelLog)
	footer := m.viewFooter()

	full := lipgloss.JoinVertical(lipgloss.Left, header, body, selectionBar, logView, footer)

	if m.modal != nil {
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center,
			m.modal.View(m.width),
			lipgloss.WithWhitespaceChars(" "),
		)
	}

	return full
}

func (m AppModel) viewHeader() string {
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
			label = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("86")).Render(label)
		}
		parts = append(parts, label)
	}

	title := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("212")).Render("pkgsh")
	return lipgloss.NewStyle().
		Width(m.width).
		Padding(0, 1).
		Background(lipgloss.Color("237")).
		Foreground(lipgloss.Color("250")).
		Render(title + "  " + strings.Join(parts, "  "))
}

func (m AppModel) viewSearchBar(width int) string {
	query := m.state.SearchQuery
	if m.searching {
		query += "█"
	}
	return lipgloss.NewStyle().
		Width(width - 2).
		Padding(0, 1).
		Render("> Buscar: " + query)
}

func (m AppModel) viewSelectionBar(width int) string {
	sel := m.list.AllSelected(m.state.Packages)
	base := lipgloss.NewStyle().Width(width).Padding(0, 1)
	if len(sel) == 0 {
		return base.Faint(true).Render("Sin selección")
	}
	names := make([]string, len(sel))
	for i, p := range sel {
		names[i] = fmt.Sprintf("%s (%s)", p.Name, string(p.Manager))
	}
	badge := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("86")).Render(fmt.Sprintf("☒ %d", len(sel)))
	text := truncate(strings.Join(names, "  ·  "), width-10)
	return base.Render(badge + "  " + text)
}

func (m AppModel) viewFooter() string {
	hints := "[/] Buscar  [Tab] Panel  [Space] Sel  [a] Todo  [Esc] Limpiar  [d] Desinstalar  [u] Actualizar  [s] Ordenar  [q] Salir"
	return lipgloss.NewStyle().
		Width(m.width).
		Faint(true).
		Padding(0, 1).
		Render(hints)
}
