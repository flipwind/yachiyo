package module

import "charm.land/lipgloss/v2"

var headerStyle = lipgloss.NewStyle().
	Bold(true)

func (m model) headerView() string {
	return headerStyle.Render("Project Yachiyo Client")
}
