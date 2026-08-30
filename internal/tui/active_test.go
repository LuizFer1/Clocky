package tui

import (
	"testing"
	"time"

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
