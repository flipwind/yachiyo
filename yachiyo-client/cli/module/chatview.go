package module

import (
	"charm.land/lipgloss/v2"
)

var chatStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		Padding(0, 1)

func (m *model) chatView() string {
	return chatStyle.Render(m.viewport.View())
}