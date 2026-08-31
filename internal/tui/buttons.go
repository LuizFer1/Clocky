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
	actionOpenPomodoro
	actionStopPomodoro
)

type actionButton struct {
	ID    actionID
	Label string
}

var (
	styleBtn = lipgloss.NewStyle().
			Foreground(colText).
			Background(colSurface2).
			Padding(0, 1).
			MarginRight(2)
	styleBtnActive = lipgloss.NewStyle().
			Foreground(lipgloss.Color("232")).
			Background(colAccent).
			Bold(true).
			Padding(0, 1).
			MarginRight(2)
	styleBtnIdle = lipgloss.NewStyle().
			Foreground(colMuted).
			Background(colSurface).
			Padding(0, 1).
			MarginRight(2)
)

// hubActionButtons builds the action row. Stopwatch label flips with state;
// Stop Timer only appears while a background timer is running.
// When a pomodoro is live, New Pomodoro is replaced by Open/Stop.
func hubActionButtons(a activeSnapshot) []actionButton {
	swLabel := "Start Stopwatch"
	if a.StopwatchRunning {
		swLabel = "Stop Stopwatch"
	}
	var btns []actionButton
	if a.PomodoroActive {
		btns = append(btns,
			actionButton{actionOpenPomodoro, "Open Pomodoro"},
			actionButton{actionStopPomodoro, "Stop Pomodoro"},
		)
	} else {
		btns = append(btns, actionButton{actionNewPomodoro, "New Pomodoro"})
	}
	btns = append(btns,
		actionButton{actionNewTimer, "New Timer"},
		actionButton{actionStopwatchToggle, swLabel},
	)
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
		label := a.Label
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
	return panelBoxFocused("Actions", row, width, actionsFocused)
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
	for i, a := range buttons {
		label := a.Label
		if actionsFocused && i == selected {
			lines = append(lines, styleBtnActive.Render(label))
		} else if actionsFocused {
			lines = append(lines, styleBtn.Render(label))
		} else {
			lines = append(lines, styleBtnIdle.Render(label))
		}
	}
	return panelBoxFocused("Actions", strings.Join(lines, "\n"), panelW, actionsFocused)
}
