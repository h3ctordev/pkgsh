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
	selected map[int]bool
}

func newListModel() ListModel {
	return ListModel{selected: make(map[int]bool)}
}

func (lm ListModel) Cursor() int { return lm.cursor }

func (lm ListModel) SetItems(pkgs []domain.Package) ListModel {
	lm.items = pkgs
	lm.cursor = 0
	lm.selected = make(map[int]bool)
	return lm
}

func (lm ListModel) ToggleSelected() ListModel {
	if len(lm.items) == 0 {
		return lm
	}
	sel := make(map[int]bool, len(lm.selected))
	for k, v := range lm.selected {
		sel[k] = v
	}
	sel[lm.cursor] = !sel[lm.cursor]
	lm.selected = sel
	return lm
}

func (lm ListModel) SelectAll() ListModel {
	sel := make(map[int]bool, len(lm.items))
	for i := range lm.items {
		sel[i] = true
	}
	lm.selected = sel
	return lm
}

func (lm ListModel) ClearSelection() ListModel {
	lm.selected = make(map[int]bool)
	return lm
}

func (lm ListModel) SelectedPackages() []domain.Package {
	var out []domain.Package
	for i, pkg := range lm.items {
		if lm.selected[i] {
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
		if lm.selected[i] {
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
