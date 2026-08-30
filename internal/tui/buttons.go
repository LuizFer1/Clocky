package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

type hubFocus int

const (
	focusActions hubFocus = iota
	focusPresets
)

type actionID int

const (
	actionNewPomodoro actionID = iota
	actionNewTimer
	actionStart
	actionStop
	actionStopwatch
)

type actionButton struct {
	ID    actionID
	Label string
}

var hubActions = []actionButton{
	{actionNewPomodoro, "New Pomodoro"},
	{actionNewTimer, "New Timer"},
	{actionStart, "Start"},
	{actionStop, "Stop"},
	{actionStopwatch, "Stopwatch"},
}

var (
	styleBtn = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252")).
			Background(lipgloss.Color("237")).
			Padding(0, 1).
			MarginRight(1)
	styleBtnActive = lipgloss.NewStyle().
			Foreground(lipgloss.Color("232")).
			Background(lipgloss.Color("51")).
			Bold(true).
			Padding(0, 1).
			MarginRight(1)
	styleBtnIdle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("245")).
			Background(lipgloss.Color("236")).
			Padding(0, 1).
			MarginRight(1)
)

func renderActionBar(selected int, actionsFocused bool, width int) string {
	var parts []string
	for i, a := range hubActions {
		label := "[ " + a.Label + " ]"
		var s string
		switch {
		case actionsFocused && i == selected:
			s = styleBtnActive.Render(label)
		case actionsFocused:
			s = styleBtn.Render(label)
		default:
			s = styleBtnIdle.Render(label)
		}
		parts = append(parts, s)
	}
	row := lipgloss.JoinHorizontal(lipgloss.Top, parts...)
	title := styleTitle.Render("Actions")
	hint := styleMuted.Render("tab switch · ←→ buttons · enter")
	inner := title + "\n" + row + "\n" + hint
	panelW := width
	if panelW < 20 {
		panelW = 20
	}
	return stylePanel.Width(panelW).Render(inner)
}

func wrapActionBar(selected int, actionsFocused bool, panelW, termW int) string {
	bar := renderActionBar(selected, actionsFocused, panelW)
	// If too wide for panel, stack buttons vertically.
	if lipgloss.Width(bar) > termW-2 {
		var lines []string
		lines = append(lines, styleTitle.Render("Actions"))
		for i, a := range hubActions {
			label := "[ " + a.Label + " ]"
			if actionsFocused && i == selected {
				lines = append(lines, styleBtnActive.Render(label))
			} else if actionsFocused {
				lines = append(lines, styleBtn.Render(label))
			} else {
				lines = append(lines, styleBtnIdle.Render(label))
			}
		}
		lines = append(lines, styleMuted.Render("tab · ↑↓ · enter"))
		return stylePanel.Width(panelW).Render(strings.Join(lines, "\n"))
	}
	return bar
}
