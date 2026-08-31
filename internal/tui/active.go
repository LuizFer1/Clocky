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

	PomodoroActive    bool
	PomodoroPhase     string
	PomodoroCycle     int
	PomodoroCycles    int
	PomodoroRemaining time.Duration
	PomodoroWaiting   bool
	PomodoroPaused    bool
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

func mergePomodoroActive(a activeSnapshot, s sessionModel) activeSnapshot {
	if !s.live() {
		return a
	}
	a.PomodoroActive = true
	if len(s.phases) == 0 || s.index >= len(s.phases) {
		return a
	}
	ph := s.phases[s.index]
	a.PomodoroPhase = ph.Name
	a.PomodoroCycle = ph.Cycle
	a.PomodoroCycles = s.cfg.Cycles
	a.PomodoroRemaining = s.remaining
	a.PomodoroWaiting = s.waitingEnter
	a.PomodoroPaused = s.paused
	return a
}
