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
	selected map[string]bool
}

func pkgKey(p domain.Package) string {
	return p.Name + ":" + string(p.Manager)
}

func newListModel() ListModel {
	return ListModel{selected: make(map[string]bool)}
}

func (lm ListModel) Cursor() int { return lm.cursor }

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

func (lm ListModel) SelectedPackages() []domain.Package {
	var out []domain.Package
	for _, pkg := range lm.items {
		if lm.selected[pkgKey(pkg)] {
			out = append(out, pkg)
		}
	}
	return out
}

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

func (lm ListModel) JumpToFirst() ListModel {
	lm.cursor = 0
	return lm
}

func (lm ListModel) JumpToLast() ListModel {
	if len(lm.items) > 0 {
		lm.cursor = len(lm.items) - 1
	}
	return lm
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
	case tea.KeyRunes:
		switch string(keyMsg.Runes) {
		case "j":
			if lm.cursor < len(lm.items)-1 {
				lm.cursor++
			}
		case "k":
			if lm.cursor > 0 {
				lm.cursor--
			}
		}
	}
	return lm, nil
}

// pkgTags devuelve los tags renderizados con color para la columna Estado.
func pkgTags(p domain.Package) string {
	var parts []string
	if p.NewVersion != "" {
		parts = append(parts, lipgloss.NewStyle().Bold(true).Foreground(colorYellow).Render("UPD"))
	}
	if domain.IsSystemPackage(p) {
		parts = append(parts, lipgloss.NewStyle().Foreground(colorOrange).Render("SYS"))
	}
	if p.IsOrphan {
		parts = append(parts, lipgloss.NewStyle().Foreground(colorRed).Render("ORP"))
	}
	return strings.Join(parts, " ")
}

func (lm ListModel) View(width, height int, active bool) string {
	style := listPanelStyle(active, width, height)
	if len(lm.items) == 0 {
		return style.Render(lipgloss.NewStyle().Faint(true).Render("Sin paquetes"))
	}

	// inner width = panel width - 2 border - 2 padding
	innerWidth := width - 6
	if innerWidth < 20 {
		innerWidth = 20
	}

	// Anchos fijos: cursor(2) checkbox(2) sep(2) version(10) sep(2) manager(8) sep(2) size(7) sep(2) tags(7) = 44
	// nombre = innerWidth - 44, mínimo 10
	fixedCols := 44
	colName := innerWidth - fixedCols
	if colName < 10 {
		colName = 10
	}
	colVer := 10
	colMgr := 8
	colSize := 7

	muted := lipgloss.NewStyle().Foreground(colorMuted)
	header := fmt.Sprintf("  %-*s  %-*s  %-*s  %-*s  %s",
		colName+4, "Paquete",
		colVer, "Versión",
		colMgr, "Gestor",
		colSize, "Tamaño",
		"Est.",
	)
	sep := strings.Repeat("─", innerWidth)
	rows := []string{
		muted.Render(header),
		muted.Render(sep),
	}

	// -2 header+sep
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

	for i := start; i < end; i++ {
		pkg := lm.items[i]
		checkbox := "☐"
		if lm.selected[pkgKey(pkg)] {
			checkbox = lipgloss.NewStyle().Foreground(colorGreen).Render("☒")
		}
		cur := " "
		if i == lm.cursor {
			cur = ">"
		}

		nameStr := truncate(pkg.Name, colName)
		mgrStr := lipgloss.NewStyle().Foreground(managerColor(pkg.Manager)).Render(
			truncate(string(pkg.Manager), colMgr),
		)
		sizeStr := truncate(formatSize(pkg.Size), colSize)
		tagsStr := pkgTags(pkg)

		row := fmt.Sprintf("%s %s %-*s  %-*s  %-*s  %-*s  %s",
			cur, checkbox,
			colName, nameStr,
			colVer, truncate(pkg.Version, colVer),
			colMgr, mgrStr,
			colSize, sizeStr,
			tagsStr,
		)

		if i == lm.cursor {
			row = lipgloss.NewStyle().
				Background(colorCursorBg).
				Foreground(colorPrimary).
				Bold(true).
				Render(row)
		}
		rows = append(rows, row)
	}

	remaining := len(lm.items) - end
	if remaining > 0 {
		rows = append(rows, muted.Render(fmt.Sprintf("  ▼ %d more", remaining)))
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
		s = s.BorderForeground(colorPrimary)
	} else {
		s = s.BorderForeground(colorBorder)
	}
	return s
}
