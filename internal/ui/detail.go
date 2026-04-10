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
		return style.Render(lipgloss.NewStyle().Faint(true).Render("Selecciona un paquete"))
	}

	label := lipgloss.NewStyle().Bold(true).Render

	newVer := pkg.NewVersion
	if newVer == "" {
		newVer = "(ninguna)"
	}
	native := "No"
	if pkg.IsNative {
		native = "Sí"
	}

	path := pkg.Path
	if path == "" {
		path = "(desconocido)"
	}
	lines := []string{
		label("Nombre:      ") + pkg.Name,
		label("Versión:     ") + pkg.Version,
		label("Nueva vers.: ") + newVer,
		label("Gestor:      ") + string(pkg.Manager),
		label("Tamaño:      ") + formatSize(pkg.Size),
		label("Nativo:      ") + native,
		label("Origen:      ") + pkg.Origin,
		label("Ruta:        ") + path,
		"",
		label("Descripción:"),
		pkg.Description,
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
		s = s.BorderForeground(lipgloss.Color("86"))
	}
	return s
}
