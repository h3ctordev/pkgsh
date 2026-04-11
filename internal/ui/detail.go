package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/hbustos/pkgsh/internal/domain"
)

type DetailModel struct{}

func newDetailModel() DetailModel { return DetailModel{} }

func (dm DetailModel) View(pkg *domain.Package, width, height int, active bool) string {
	style := detailPanelStyle(active, width, height)
	if pkg == nil {
		return style.Render(lipgloss.NewStyle().Foreground(colorMuted).Render("Selecciona un paquete"))
	}

	label := lipgloss.NewStyle().Bold(true).Foreground(colorPrimary).Render

	verDisplay := pkg.Version
	if pkg.NewVersion != "" {
		verDisplay = pkg.Version + " → " +
			lipgloss.NewStyle().Foreground(colorYellow).Render(pkg.NewVersion)
	}

	mgrDisplay := lipgloss.NewStyle().Foreground(managerColor(pkg.Manager)).Render(string(pkg.Manager))

	native := "No"
	if pkg.IsNative {
		native = "Sí"
	}
	path := pkg.Path
	if path == "" {
		path = "(desconocido)"
	}

	lines := []string{
		label("Nombre:  ") + pkg.Name,
		label("Versión: ") + verDisplay,
		label("Gestor:  ") + mgrDisplay,
		label("Tamaño:  ") + formatSize(pkg.Size),
		label("Nativo:  ") + native,
		label("Origen:  ") + pkg.Origin,
		label("Ruta:    ") + truncate(path, width-14),
		"",
		label("Descripción:"),
	}

	// Líneas disponibles para descripción
	descAvail := height - len(lines) - 3
	if descAvail < 1 {
		descAvail = 1
	}
	descLines := strings.Split(pkg.Description, "\n")
	for i, l := range descLines {
		if i >= descAvail {
			lines = append(lines, lipgloss.NewStyle().Foreground(colorMuted).Render("…"))
			break
		}
		lines = append(lines, lipgloss.NewStyle().Foreground(colorMuted).Render(truncate(l, width-6)))
	}

	return style.Render(strings.Join(lines, "\n"))
}

func formatSize(bytes int64) string {
	switch {
	case bytes >= 1024*1024*1024:
		return fmt.Sprintf("%.1f GB", float64(bytes)/(1024*1024*1024))
	case bytes >= 1024*1024:
		return fmt.Sprintf("%.1f MB", float64(bytes)/(1024*1024))
	case bytes >= 1024:
		return fmt.Sprintf("%.1f KB", float64(bytes)/1024)
	default:
		if bytes == 0 {
			return "—"
		}
		return fmt.Sprintf("%d B", bytes)
	}
}

func detailPanelStyle(active bool, width, height int) lipgloss.Style {
	s := lipgloss.NewStyle().
		Width(width - 2).
		Height(height).
		Border(lipgloss.RoundedBorder()).
		Padding(1, 2)
	if active {
		s = s.BorderForeground(colorPrimary)
	} else {
		s = s.BorderForeground(colorBorder)
	}
	return s
}
