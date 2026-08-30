package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

type confirmResultMsg struct {
	OK bool
}

type confirmModel struct {
	kind   string
	name   string
	width  int
	height int
}

func newConfirmModel(kind, name string) confirmModel {
	return confirmModel{kind: kind, name: name, width: 80, height: 24}
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
	w, ht := c.width, c.height
	if w <= 0 {
		w = 80
	}
	if ht <= 0 {
		ht = 24
	}
	panelW := min(48, max(28, w-10))
	msg := fmt.Sprintf("Delete %s preset %q?", c.kind, c.name)
	body := panelBox("Confirm delete", msg, panelW)
	return fillFrame(w, ht, "Clocky", body, "y confirm  n/esc cancel")
}
