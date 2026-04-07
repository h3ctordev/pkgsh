package ui

import (
	"bufio"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/hbustos/pkgsh/internal/domain"
)

// operationLineMsg arrives each time a line is read from the active operation.
type operationLineMsg string

// operationDoneMsg arrives when the active operation finishes (err may be nil).
type operationDoneMsg struct{ err error }

type LogModel struct {
	lines     []string
	scrollOff int
}

func newLogModel() LogModel {
	return LogModel{}
}

func (lm LogModel) appendLine(line string) LogModel {
	lm.lines = append(lm.lines, line)
	return lm
}

// Update handles PgUp/PgDn when the log panel is active.
func (lm LogModel) Update(msg tea.Msg) (LogModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyPgUp:
			if lm.scrollOff > 0 {
				lm.scrollOff -= 5
				if lm.scrollOff < 0 {
					lm.scrollOff = 0
				}
			}
		case tea.KeyPgDown:
			lm.scrollOff += 5
		}
	}
	return lm, nil
}

// View renders the log panel showing the most recent visible lines.
func (lm LogModel) View(width, height int, active bool) string {
	style := logPanelStyle(active, width, height)
	if len(lm.lines) == 0 {
		return style.Render(lipgloss.NewStyle().Faint(true).Render("Sin operaciones"))
	}

	// Calculate visible window
	start := len(lm.lines) - height - lm.scrollOff
	if start < 0 {
		start = 0
	}
	end := start + height
	if end > len(lm.lines) {
		end = len(lm.lines)
	}

	var rendered []string
	for _, line := range lm.lines[start:end] {
		rendered = append(rendered, line)
	}
	return style.Render(lipgloss.JoinVertical(lipgloss.Left, rendered...))
}

// readLineCmd returns a tea.Cmd that reads the next line from the active operation.
// Must be called recursively until operationDoneMsg is received.
func readLineCmd(op *domain.Operation) tea.Cmd {
	return func() tea.Msg {
		scanner := bufio.NewScanner(op.Reader())
		if scanner.Scan() {
			return operationLineMsg(scanner.Text())
		}
		return operationDoneMsg{err: scanner.Err()}
	}
}

func logPanelStyle(active bool, width, height int) lipgloss.Style {
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
