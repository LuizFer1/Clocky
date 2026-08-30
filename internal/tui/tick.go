package tui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type tickMsg struct{}

func scheduleTick() tea.Cmd {
	return tea.Tick(time.Second, func(time.Time) tea.Msg {
		return tickMsg{}
	})
}
