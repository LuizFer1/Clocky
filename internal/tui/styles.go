package tui

import "github.com/charmbracelet/lipgloss"

var (
	styleTitle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39"))
	styleMuted = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
	styleError = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	styleOK    = lipgloss.NewStyle().Foreground(lipgloss.Color("82"))
	styleSel   = lipgloss.NewStyle().Foreground(lipgloss.Color("229")).Bold(true)
	stylePanel = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("63")).
			Padding(0, 1)
	stylePaused = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("214"))
)
