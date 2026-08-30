package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type confirmResultMsg struct {
	OK bool
}

type confirmModel struct {
	kind string
	name string
}

func newConfirmModel(kind, name string) confirmModel {
	return confirmModel{kind: kind, name: name}
}

func (c confirmModel) Init() tea.Cmd { return nil }

func (c confirmModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "y", "Y":
			return c, func() tea.Msg { return confirmResultMsg{OK: true} }
		case "n", "N", "esc":
			return c, func() tea.Msg { return confirmResultMsg{OK: false} }
		case "q", "ctrl+c":
			return c, tea.Quit
		}
	}
	return c, nil
}

func (c confirmModel) View() string {
	var b strings.Builder
	b.WriteString(styleTitle.Render("Confirm delete"))
	b.WriteString("\n\n")
	b.WriteString(fmt.Sprintf("Delete %s preset %q?\n", c.kind, c.name))
	b.WriteString("\n")
	b.WriteString(styleMuted.Render("y confirm  n/esc cancel"))
	return b.String()
}
