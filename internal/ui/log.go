package ui

import (
	"bufio"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/hbustos/pkgsh/internal/domain"
)

type operationLineMsg string
type operationDoneMsg struct{ err error }

type LogModel struct {
	lines      []string
	scrollOff  int
	autoScroll bool
}

func newLogModel() LogModel {
	return LogModel{autoScroll: true}
}

func (lm LogModel) appendLine(line string) LogModel {
	// dpkg uses \r to overwrite progress in-place; keep only the last segment.
	if i := strings.LastIndex(line, "\r"); i >= 0 {
		line = line[i+1:]
	}
	lm.lines = append(lm.lines, line)
	if lm.autoScroll {
		lm.scrollOff = 0
	}
	return lm
}

func (lm LogModel) Update(msg tea.Msg) (LogModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyPgUp:
			lm.autoScroll = false
			if lm.scrollOff < len(lm.lines) {
				lm.scrollOff += 5
			}
		case tea.KeyPgDown:
			if lm.scrollOff > 0 {
				lm.scrollOff -= 5
				if lm.scrollOff < 0 {
					lm.scrollOff = 0
				}
			} else {
				lm.autoScroll = true
			}
		}
	}
	return lm, nil
}

// colorLine aplica color semántico a una línea de log.
func colorLine(line string) string {
	switch {
	case strings.Contains(line, "[ERROR]"):
		return lipgloss.NewStyle().Foreground(colorRed).Render(line)
	case strings.Contains(line, "[SECURITY]"):
		return lipgloss.NewStyle().Foreground(colorSec).Render(line)
	case strings.Contains(line, "[SKIP]"):
		return lipgloss.NewStyle().Faint(true).Render(line)
	case strings.HasPrefix(line, "Removing ") ||
		strings.HasPrefix(line, "Unpacking ") ||
		strings.HasPrefix(line, "Setting up "):
		return lipgloss.NewStyle().Foreground(colorGreen).Render(line)
	default:
		return lipgloss.NewStyle().Foreground(colorMuted).Render(line)
	}
}

// View renderiza el panel de log.
// title: cadena vacía = log colapsado (muestra 1 línea estática).
// title != "" = log activo con título de operación en la primera fila.
func (lm LogModel) View(width, height int, active bool, title string) string {
	if height <= 1 {
		msg := lipgloss.NewStyle().Foreground(colorMuted).Render("Sin operaciones activas")
		return lipgloss.NewStyle().
			Width(width - 2).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorBorder).
			Padding(0, 1).
			Render(msg)
	}

	style := logPanelStyle(active, width, height)
	if len(lm.lines) == 0 && title == "" {
		return style.Render(lipgloss.NewStyle().Foreground(colorMuted).Render("Sin operaciones activas"))
	}

	// Si hay título (operación activa), reservar 2 filas: título + separador
	contentHeight := height
	var header []string
	if title != "" {
		titleStyled := lipgloss.NewStyle().Foreground(colorGreen).Bold(true).Render(title)
		sep := lipgloss.NewStyle().Foreground(colorBorder).Render(strings.Repeat("─", width-6))
		header = []string{titleStyled, sep}
		contentHeight = height - 2
		if contentHeight < 1 {
			contentHeight = 1
		}
	}

	start := len(lm.lines) - contentHeight - lm.scrollOff
	if start < 0 {
		start = 0
	}
	end := start + contentHeight
	if end > len(lm.lines) {
		end = len(lm.lines)
	}

	// inner width = panel width - 2 (border) - 2 (padding) - 1 (margin)
	maxLineWidth := width - 5
	if maxLineWidth < 10 {
		maxLineWidth = 10
	}

	var rendered []string
	rendered = append(rendered, header...)
	for _, line := range lm.lines[start:end] {
		rendered = append(rendered, colorLine(truncate(line, maxLineWidth)))
	}

	return style.Render(lipgloss.JoinVertical(lipgloss.Left, rendered...))
}

func readLineCmd(op *domain.Operation) tea.Cmd {
	return func() tea.Msg {
		scanner := bufio.NewScanner(op.Reader())
		if scanner.Scan() {
			return operationLineMsg(scanner.Text())
		}
		return operationDoneMsg{err: scanner.Err()}
	}
}

func (lm LogModel) Lines() []string { return lm.lines }

func logPanelStyle(active bool, width, height int) lipgloss.Style {
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
