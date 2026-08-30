package stopwatch

import (
	"errors"
	"os"
	"time"

	"github.com/LuizFer1/Clocky/internal/state"
)

// State is the on-disk stopwatch marker.
type State struct {
	StartedAt time.Time `json:"started_at"`
}

// Toggle starts a stopwatch when none is running, or stops and returns elapsed.
func Toggle(root string, now time.Time) (action string, elapsed time.Duration, err error) {
	path := state.Path(root, "stopwatch.json")
	var st State
	err = state.ReadJSON(path, &st)
	if errors.Is(err, os.ErrNotExist) {
		if err := state.EnsureDir(root); err != nil {
			return "", 0, err
		}
		st = State{StartedAt: now}
		if err := state.WriteJSON(path, st); err != nil {
			return "", 0, err
		}
		return "started", 0, nil
	}
	if err != nil {
		return "", 0, err
	}
	elapsed = now.Sub(st.StartedAt)
	if err := os.Remove(path); err != nil {
		return "", 0, err
	}
	return "stopped", elapsed, nil
}

// Status reports whether a stopwatch is running and its elapsed time.
func Status(root string, now time.Time) (running bool, elapsed time.Duration, err error) {
	path := state.Path(root, "stopwatch.json")
	var st State
	err = state.ReadJSON(path, &st)
	if errors.Is(err, os.ErrNotExist) {
		return false, 0, nil
	}
	if err != nil {
		return false, 0, err
	}
	return true, now.Sub(st.StartedAt), nil
}
