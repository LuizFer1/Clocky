package tui

import "github.com/charmbracelet/lipgloss"

// Palette — one place for every colour the TUI uses.
var (
	colAccent   = lipgloss.Color("51")  // cyan — brand / focus
	colAccent2  = lipgloss.Color("219") // pink — phase / highlight
	colWarn     = lipgloss.Color("214") // amber — paused / long break
	colOK       = lipgloss.Color("114") // green — running / success
	colErr      = lipgloss.Color("203") // red — errors
	colText     = lipgloss.Color("252")
	colMuted    = lipgloss.Color("245")
	colDim      = lipgloss.Color("240")
	colBorder   = lipgloss.Color("60")
	colBorderF  = lipgloss.Color("51")
	colSurface  = lipgloss.Color("236")
	colSurface2 = lipgloss.Color("238")
)

var (
	styleTitle  = lipgloss.NewStyle().Bold(true).Foreground(colAccent)
	styleMuted  = lipgloss.NewStyle().Foreground(colMuted)
	styleDim    = lipgloss.NewStyle().Foreground(colDim)
	styleError  = lipgloss.NewStyle().Foreground(colErr).Bold(true)
	styleOK     = lipgloss.NewStyle().Foreground(colOK)
	styleSel    = lipgloss.NewStyle().Foreground(lipgloss.Color("229")).Bold(true).Background(colSurface2)
	stylePaused = lipgloss.NewStyle().Bold(true).Foreground(colWarn)
	stylePhase  = lipgloss.NewStyle().Bold(true).Foreground(colAccent2)
	styleKey    = lipgloss.NewStyle().Bold(true).Foreground(colAccent)

	stylePanel = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colBorder).
			Padding(0, 2)
	stylePanelFocused = stylePanel.BorderForeground(colBorderF)
)
