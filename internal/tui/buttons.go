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
	actionStopwatchToggle
	actionStopTimer
)

type actionButton struct {
	ID    actionID
	Label string
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

// hubActionButtons builds the action row. Stopwatch label flips with state;
// Stop Timer only appears while a background timer is running.
func hubActionButtons(a activeSnapshot) []actionButton {
	swLabel := "Start Stopwatch"
	if a.StopwatchRunning {
		swLabel = "Stop Stopwatch"
	}
	btns := []actionButton{
		{actionNewPomodoro, "New Pomodoro"},
		{actionNewTimer, "New Timer"},
		{actionStopwatchToggle, swLabel},
	}
	if a.TimerActive {
		btns = append(btns, actionButton{actionStopTimer, "Stop Timer"})
	}
	return btns
}

func renderActionBar(buttons []actionButton, selected int, actionsFocused bool, width int) string {
	if len(buttons) == 0 {
		return ""
	}
	if selected >= len(buttons) {
		selected = len(buttons) - 1
	}
	if selected < 0 {
		selected = 0
	}
	var parts []string
	for i, a := range buttons {
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

func wrapActionBar(buttons []actionButton, selected int, actionsFocused bool, panelW, termW int) string {
	bar := renderActionBar(buttons, selected, actionsFocused, panelW)
	if lipgloss.Width(bar) <= termW-2 {
		return bar
	}
	if selected >= len(buttons) {
		selected = len(buttons) - 1
	}
	var lines []string
	lines = append(lines, styleTitle.Render("Actions"))
	for i, a := range buttons {
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
