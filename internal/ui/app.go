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

type AppModel struct {
	state     domain.AppState
	list      ListModel
	detail    DetailModel
	log       LogModel
	modal     *ModalModel
	searching bool
	width     int
	height    int
}

func New(pkgs []domain.Package, opts Options) AppModel {
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

	return AppModel{
		state:  state,
		list:   newListModel().SetItems(filtered),
		detail: newDetailModel(),
		log:    newLogModel(),
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

	case operationLineMsg:
		m.log = m.log.appendLine(string(msg))
		return m, readLineCmd(m.state.Operation)

	case operationDoneMsg:
		m.state.Operation = nil
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
			m.log = m.log.appendLine("Operación iniciada (mock)")
			m.list = m.list.ClearSelection()
			m.modal = nil
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
