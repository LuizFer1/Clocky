package tui

import (
	"testing"
	"time"

	"github.com/LuizFer1/Clocky/internal/pomodoro"
	"github.com/LuizFer1/Clocky/internal/stopwatch"
)

func TestRefreshActiveReadsStopwatch(t *testing.T) {
	root := t.TempDir()
	if _, _, err := stopwatch.Toggle(root, time.Unix(1000, 0)); err != nil {
		t.Fatal(err)
	}
	a := refreshActive(root, time.Unix(1010, 0))
	if !a.StopwatchRunning {
		t.Fatal("expected stopwatch running")
	}
	if a.StopwatchElapsed != 10*time.Second {
		t.Fatalf("elapsed=%v", a.StopwatchElapsed)
	}
}

func TestRefreshActiveIdleWhenEmpty(t *testing.T) {
	a := refreshActive(t.TempDir(), time.Now())
	if a.TimerActive || a.StopwatchRunning {
		t.Fatalf("expected idle: %+v", a)
	}
}

func TestMergePomodoroActiveFromLiveSession(t *testing.T) {
	cfg := pomodoro.Config{
		Focus:  25 * time.Minute,
		Break:  5 * time.Minute,
		Long:   15 * time.Minute,
		Cycles: 4,
		Auto:   true,
	}
	s := newSessionModel(cfg, 40, 24, nil)
	if !s.live() {
		t.Fatal("expected live session")
	}
	a := mergePomodoroActive(activeSnapshot{}, s)
	if !a.PomodoroActive || a.PomodoroPhase != "FOCUS" || a.PomodoroCycle != 1 || a.PomodoroCycles != 4 {
		t.Fatalf("got %+v", a)
	}
	if a.PomodoroRemaining != 25*time.Minute {
		t.Fatalf("remaining=%v", a.PomodoroRemaining)
	}
}

func TestMergePomodoroActiveIgnoresInactive(t *testing.T) {
	var s sessionModel
	if s.live() {
		t.Fatal("empty session should not be live")
	}
	a := mergePomodoroActive(activeSnapshot{}, s)
	if a.PomodoroActive {
		t.Fatal("expected inactive")
	}
}

func TestMergePomodoroActiveIgnoresOutOfRangeIndex(t *testing.T) {
	s := sessionModel{
		active: true,
		phases: []pomodoro.Phase{{Name: "FOCUS", Cycle: 1, Duration: time.Minute}},
		index:  1,
	}
	if !s.live() {
		t.Fatal("expected live session")
	}
	a := mergePomodoroActive(activeSnapshot{}, s)
	if a.PomodoroActive {
		t.Fatalf("expected inactive for out-of-range index, got %+v", a)
	}
}
