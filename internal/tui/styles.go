package tui

import "github.com/charmbracelet/lipgloss"

var (
	styleTitle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("51"))
	styleMuted = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
	styleError = lipgloss.NewStyle().Foreground(lipgloss.Color("203")).Bold(true)
	styleOK    = lipgloss.NewStyle().Foreground(lipgloss.Color("114"))
	styleSel   = lipgloss.NewStyle().Foreground(lipgloss.Color("229")).Bold(true).Background(lipgloss.Color("238"))
	stylePanel = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("62")).
			Padding(1, 2).
			Margin(0, 0, 1, 0)
	stylePaused = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("214"))
	stylePhase  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("219"))
)
