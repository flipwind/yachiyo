package module

import "charm.land/lipgloss/v2"

var textinputStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		Padding(0, 1)

func (m model) textInputView() string {
	return textinputStyle.Render(m.textInput.View())
}