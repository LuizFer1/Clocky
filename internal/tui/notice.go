package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// timerFinishedNotice returns a hub notice when a running timer reaches zero
// or disappears after having remaining time (worker cleared state).
// Manual stops set status to "Timer stopped" and should not use this.
func timerFinishedNotice(prev, cur activeSnapshot, currentStatus string) (string, bool) {
	if currentStatus == "Timer stopped" {
		return "", false
	}
	if !prev.TimerActive || prev.TimerRemaining <= 0 {
		return "", false
	}
	crossedZero := cur.TimerActive && cur.TimerRemaining == 0
	cleared := !cur.TimerActive
	if !crossedZero && !cleared {
		return "", false
	}
	label := prev.TimerLabel
	if label == "" {
		label = "timer"
	}
	return fmt.Sprintf("Timer finished: %s", label), true
}

func renderNotice(msg string, width int) string {
	if strings.TrimSpace(msg) == "" {
		return ""
	}
	style := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("232")).
		Background(lipgloss.Color("214")).
		Padding(0, 1)
	inner := style.Render("! "+msg) + "\n" + styleMuted.Render("press esc to dismiss")
	return panelBox("Notice", inner, width)
}
