package tui

import (
	"time"

	"github.com/LuizFer1/Clocky/internal/stopwatch"
	"github.com/LuizFer1/Clocky/internal/timer"
)

type activeSnapshot struct {
	TimerActive      bool
	TimerRemaining   time.Duration
	TimerLabel       string
	StopwatchRunning bool
	StopwatchElapsed time.Duration
}

func refreshActive(root string, now time.Time) activeSnapshot {
	var a activeSnapshot
	if active, rem, label, err := timer.Status(root, now); err == nil {
		a.TimerActive = active
		a.TimerRemaining = rem
		a.TimerLabel = label
	}
	if running, elapsed, err := stopwatch.Status(root, now); err == nil {
		a.StopwatchRunning = running
		a.StopwatchElapsed = elapsed
	}
	return a
}
