package tui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type tickMsg struct{}
type alertTickMsg struct{}

func scheduleTick() tea.Cmd {
	return tea.Tick(time.Second, func(time.Time) tea.Msg {
		return tickMsg{}
	})
}

// scheduleAlertTick repeats the hub alarm while a finished-timer notice is active.
func scheduleAlertTick() tea.Cmd {
	return tea.Tick(1250*time.Millisecond, func(time.Time) tea.Msg {
		return alertTickMsg{}
	})
}
