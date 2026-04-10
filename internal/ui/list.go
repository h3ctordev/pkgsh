package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/hbustos/pkgsh/internal/domain"
)

type ListModel struct {
	items    []domain.Package
	cursor   int
	selected map[string]bool // clave: "name:manager" — persiste entre filtros
}

// pkgKey devuelve la clave de identidad de un paquete.
func pkgKey(p domain.Package) string {
	return p.Name + ":" + string(p.Manager)
}

func newListModel() ListModel {
	return ListModel{selected: make(map[string]bool)}
}

func (lm ListModel) Cursor() int { return lm.cursor }

// SetItems actualiza los items visibles y resetea el cursor.
// La selección se preserva — las claves por nombre+gestor sobreviven al filtro.
func (lm ListModel) SetItems(pkgs []domain.Package) ListModel {
	lm.items = pkgs
	lm.cursor = 0
	return lm
}

func (lm ListModel) ToggleSelected() ListModel {
	if len(lm.items) == 0 {
		return lm
	}
	sel := make(map[string]bool, len(lm.selected))
	for k, v := range lm.selected {
		sel[k] = v
	}
	key := pkgKey(lm.items[lm.cursor])
	sel[key] = !sel[key]
	lm.selected = sel
	return lm
}

func (lm ListModel) SelectAll() ListModel {
	sel := make(map[string]bool, len(lm.selected)+len(lm.items))
	for k, v := range lm.selected {
		sel[k] = v
	}
	for _, p := range lm.items {
		sel[pkgKey(p)] = true
	}
	lm.selected = sel
	return lm
}

func (lm ListModel) ClearSelection() ListModel {
	lm.selected = make(map[string]bool)
	return lm
}

// SelectedPackages devuelve los paquetes seleccionados visibles (items actuales).
func (lm ListModel) SelectedPackages() []domain.Package {
	var out []domain.Package
	for _, pkg := range lm.items {
		if lm.selected[pkgKey(pkg)] {
			out = append(out, pkg)
		}
	}
	return out
}

// AllSelected devuelve todos los paquetes seleccionados del listado completo,
// incluyendo los que no son visibles en el filtro activo.
func (lm ListModel) AllSelected(all []domain.Package) []domain.Package {
	var out []domain.Package
	for _, pkg := range all {
		if lm.selected[pkgKey(pkg)] {
			out = append(out, pkg)
		}
	}
	return out
}

func (lm ListModel) CurrentPackage() *domain.Package {
	if len(lm.items) == 0 || lm.cursor >= len(lm.items) {
		return nil
	}
	p := lm.items[lm.cursor]
	return &p
}

func (lm ListModel) Update(msg tea.Msg) (ListModel, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return lm, nil
	}
	switch keyMsg.Type {
	case tea.KeyUp:
		if lm.cursor > 0 {
			lm.cursor--
		}
	case tea.KeyDown:
		if lm.cursor < len(lm.items)-1 {
			lm.cursor++
		}
	}
	return lm, nil
}

func (lm ListModel) View(width, height int, active bool) string {
	style := listPanelStyle(active, width, height)
	if len(lm.items) == 0 {
		return style.Render(lipgloss.NewStyle().Faint(true).Render("Sin paquetes"))
	}

	// -2 bordes, -2 header+separador
	visible := height - 4
	if visible < 1 {
		visible = 1
	}

	start := 0
	if lm.cursor >= visible {
		start = lm.cursor - visible + 1
	}
	end := start + visible
	if end > len(lm.items) {
		end = len(lm.items)
	}

	// Anchos de columna: nombre=24, versión=12, gestor=10, actualizable=2
	colName := 24
	colVer := 12
	colMgr := 10

	header := fmt.Sprintf("  %-*s  %-*s  %-*s  %s",
		colName+2, "Paquete",
		colVer, "Versión",
		colMgr, "Gestor",
		"Upd",
	)
	sep := strings.Repeat("─", width-4)
	rows := []string{
		lipgloss.NewStyle().Faint(true).Render(header),
		lipgloss.NewStyle().Faint(true).Render(sep),
	}

	for i := start; i < end; i++ {
		pkg := lm.items[i]
		checkbox := "☐"
		if lm.selected[pkgKey(pkg)] {
			checkbox = "☒"
		}
		cur := " "
		if i == lm.cursor {
			cur = ">"
		}
		upd := " "
		if pkg.NewVersion != "" {
			upd = "↑"
		}
		nameCol := fmt.Sprintf("%s %s %s", cur, checkbox, truncate(pkg.Name, colName))
		row := fmt.Sprintf("%-*s  %-*s  %-*s  %s",
			colName+4, nameCol,
			colVer, truncate(pkg.Version, colVer),
			colMgr, string(pkg.Manager),
			upd,
		)
		if i == lm.cursor {
			row = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("86")).Render(row)
		}
		rows = append(rows, row)
	}

	return style.Render(strings.Join(rows, "\n"))
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-1] + "…"
}

func listPanelStyle(active bool, width, height int) lipgloss.Style {
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
