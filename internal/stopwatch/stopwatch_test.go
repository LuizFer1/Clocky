package stopwatch

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/luisf/clocky/internal/state"
)

func TestToggleStartThenStop(t *testing.T) {
	root := t.TempDir()
	now := time.Date(2026, 8, 29, 12, 0, 0, 0, time.UTC)

	action, elapsed, err := Toggle(root, now)
	if err != nil {
		t.Fatalf("Toggle start: %v", err)
	}
	if action != "started" {
		t.Fatalf("action = %q want started", action)
	}
	if elapsed != 0 {
		t.Fatalf("elapsed = %v want 0", elapsed)
	}

	path := state.Path(root, "stopwatch.json")
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected stopwatch.json after start: %v", err)
	}

	var st State
	if err := state.ReadJSON(path, &st); err != nil {
		t.Fatalf("ReadJSON: %v", err)
	}
	if !st.StartedAt.Equal(now) {
		t.Fatalf("StartedAt = %v want %v", st.StartedAt, now)
	}

	later := now.Add(5 * time.Second)
	action, elapsed, err = Toggle(root, later)
	if err != nil {
		t.Fatalf("Toggle stop: %v", err)
	}
	if action != "stopped" {
		t.Fatalf("action = %q want stopped", action)
	}
	if elapsed != 5*time.Second {
		t.Fatalf("elapsed = %v want 5s", elapsed)
	}
	if _, err := os.Stat(path); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("expected stopwatch.json removed, got %v", err)
	}
}

func TestStatusRunningAndIdle(t *testing.T) {
	root := t.TempDir()
	now := time.Date(2026, 8, 29, 12, 0, 0, 0, time.UTC)

	running, elapsed, err := Status(root, now)
	if err != nil {
		t.Fatalf("Status idle: %v", err)
	}
	if running {
		t.Fatal("expected not running")
	}
	if elapsed != 0 {
		t.Fatalf("elapsed = %v want 0", elapsed)
	}

	if _, _, err := Toggle(root, now); err != nil {
		t.Fatalf("Toggle start: %v", err)
	}

	later := now.Add(3 * time.Second)
	running, elapsed, err = Status(root, later)
	if err != nil {
		t.Fatalf("Status running: %v", err)
	}
	if !running {
		t.Fatal("expected running")
	}
	if elapsed != 3*time.Second {
		t.Fatalf("elapsed = %v want 3s", elapsed)
	}
}

func TestToggleCreatesRoot(t *testing.T) {
	root := filepath.Join(t.TempDir(), "nested", ".clocky")
	now := time.Now().UTC()
	if _, _, err := Toggle(root, now); err != nil {
		t.Fatalf("Toggle: %v", err)
	}
	if _, err := os.Stat(state.Path(root, "stopwatch.json")); err != nil {
		t.Fatalf("stat: %v", err)
	}
}
