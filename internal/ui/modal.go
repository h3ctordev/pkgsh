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
	ModalHelp
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

func newHelpModal() ModalModel {
	return ModalModel{modalType: ModalHelp}
}

// Update processes keypresses for the modal.
// Returns (updated model, confirmed, cancelled).
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

	case ModalHelp:
		switch keyMsg.String() {
		case "?", "esc":
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

// View renders the modal centered on screen.
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
	case ModalHelp:
		content = `NAVEGACIÓN
  ↑/k  Subir           ↓/j  Bajar
  g    Ir al inicio    G    Ir al final
  Tab  Cambiar panel   1-8  Filtrar por gestor

SELECCIÓN
  Space  Seleccionar/deseleccionar
  a      Seleccionar todo visible
  Esc    Limpiar selección

ACCIONES
  d  Desinstalar seleccionados
  u  Actualizar seleccionados
  r  Recargar lista

BÚSQUEDA
  /  Buscar           s  Ordenar
  S  Security mode

APP
  ?  Ayuda    q  Salir`
	}

	boxWidth := 50
	if boxWidth > width-4 {
		boxWidth = width - 4
	}

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorPrimary).
		Padding(1, 2).
		Width(boxWidth).
		Render(content)

	return lipgloss.Place(width, 10, lipgloss.Center, lipgloss.Center, box)
}
