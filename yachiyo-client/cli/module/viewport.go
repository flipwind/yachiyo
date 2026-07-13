package module

import (
	"strings"
	"charm.land/lipgloss/v2"
)

func (m *model) updateViewport() {
	var lines []string

	style := lipgloss.NewStyle().
    	Width(m.viewport.Width())

	for _, msg := range m.messages {
		rendertext := style.Render(msg.String())
		lines = append(lines, rendertext)
	}

	m.viewport.SetContent(strings.Join(lines, "\n"))
	m.viewport.GotoBottom()
}
