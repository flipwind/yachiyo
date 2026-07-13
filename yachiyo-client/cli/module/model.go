package module

import (
	"time"
	"yachiyo/yachiyo-client/cli/event"
	"yachiyo/yachiyo-client/cli/gateway"

	"charm.land/bubbles/v2/textinput"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type model struct {
	width   int
	height  int
	channel *gateway.CliChannel

	messages  []Message
	textInput textinput.Model
	viewport  viewport.Model
	quitting  bool
}

func InitialModel(cc *gateway.CliChannel) model {
	ti := textinput.New()
	ti.Placeholder = "Contacting yachiyo..."
	ti.Focus()

	vp := viewport.New()
	return model{
		channel:   cc,
		messages:  make([]Message, 0),
		textInput: ti,
		viewport:  vp,
	}
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			m.quitting = true
			return m, tea.Quit
		case "enter":
			m.messages = append(m.messages, Message{
				sender:  "User",
				time:    time.Now(),
				content: m.textInput.Value(),
			})

			m.channel.ToServer <- event.Message{
				Content: m.textInput.Value(),
			}

			m.textInput.SetValue("")
			m.updateViewport()
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		inputHeight := lipgloss.Height(m.textInputView())
		headerHeight := lipgloss.Height(m.headerView())
		footerHeight := lipgloss.Height(m.footerView())

		width := msg.Width - 4

		m.viewport.SetHeight(
			msg.Height - headerHeight - inputHeight - footerHeight - 4,
		)
		m.viewport.SetWidth(width)

		m.textInput.SetWidth(width - 3)
	case event.Message:
		m.messages = append(m.messages, Message{
			sender:  "Yachiyo",
			time:    time.Now(),
			content: msg.Content,
		})
		m.updateViewport()
	}

	m.textInput, cmd = m.textInput.Update(msg)
	cmds = append(cmds, cmd)

	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m model) View() tea.View {
	var c *tea.Cursor
	if !m.textInput.VirtualCursor() {
		c = m.textInput.Cursor()
		c.Y += lipgloss.Height(m.headerView())
	}

	str := lipgloss.JoinVertical(lipgloss.Top, m.headerView(), m.chatView(), m.textInputView(), m.footerView())
	if m.quitting {
		str += "\n"
	}

	v := tea.NewView(str)
	v.Cursor = c
	v.MouseMode = tea.MouseModeAllMotion
	return v
}
