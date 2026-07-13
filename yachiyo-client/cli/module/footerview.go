package module

import "charm.land/lipgloss/v2"

var footerStyle = lipgloss.NewStyle().
	Bold(true)

func (m model) footerView() string {
	return footerStyle.Render("(esc to exit)")
}