package ui

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/hbustos/pkgsh/internal/domain"
)

// Paleta GitHub Dark
const (
	colorPrimary  = lipgloss.Color("#58a6ff") // azul hielo — cursor, foco, activo
	colorGreen    = lipgloss.Color("#3fb950") // verde — selección marcada
	colorYellow   = lipgloss.Color("#ffd700") // amarillo — UPD tag
	colorOrange   = lipgloss.Color("#d29922") // amarillo oscuro — SYS tag
	colorRed      = lipgloss.Color("#f85149") // rojo — ORP tag, errores
	colorSec      = lipgloss.Color("#e3b341") // naranja — security mode
	colorMuted    = lipgloss.Color("#8b949e") // gris — texto secundario
	colorBg       = lipgloss.Color("#0d1117") // fondo
	colorBorder   = lipgloss.Color("#30363d") // bordes
	colorHeaderBg = lipgloss.Color("#161b22") // fondo header
	colorCursorBg = lipgloss.Color("#1f3d5c") // fondo fila activa

	// Colores por gestor
	colorApt      = lipgloss.Color("#58a6ff")
	colorSnap     = lipgloss.Color("#e3b341")
	colorFlatpak  = lipgloss.Color("#bc8cff")
	colorNpm      = lipgloss.Color("#3fb950")
	colorPip      = lipgloss.Color("#f85149")
	colorDpkg     = lipgloss.Color("#8b949e")
	colorAppImage = lipgloss.Color("#fb8f44")
)

// managerColor devuelve el color asociado a un gestor de paquetes.
func managerColor(m domain.ManagerType) lipgloss.Color {
	switch m {
	case domain.ManagerApt:
		return colorApt
	case domain.ManagerSnap:
		return colorSnap
	case domain.ManagerFlatpak:
		return colorFlatpak
	case domain.ManagerNpm:
		return colorNpm
	case domain.ManagerPip:
		return colorPip
	case domain.ManagerDpkg:
		return colorDpkg
	case domain.ManagerAppImage:
		return colorAppImage
	default:
		return colorMuted
	}
}
