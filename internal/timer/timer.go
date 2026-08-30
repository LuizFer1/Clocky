package timer

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/luisf/clocky/internal/notify"
	"github.com/luisf/clocky/internal/state"
)

const fileName = "timer.json"

// State is the on-disk background timer marker.
type State struct {
	Deadline time.Time `json:"deadline"`
	PID      int       `json:"pid"`
	Label    string    `json:"label"`
}

// Start launches a detached worker and records timer.json.
func Start(root string, d time.Duration, label string, exe string, workerArgs []string) error {
	path := state.Path(root, fileName)
	var existing State
	err := state.ReadJSON(path, &existing)
	if err == nil && processAlive(existing.PID) {
		return fmt.Errorf("timer already running")
	}
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	// Stale file is overwritten after a successful spawn.

	if err := state.EnsureDir(root); err != nil {
		return err
	}

	cmd := exec.Command(exe, workerArgs...)
	cmd.Dir = root
	detach(cmd)

	if err := cmd.Start(); err != nil {
		return err
	}
	pid := cmd.Process.Pid
	_ = cmd.Process.Release()

	st := State{
		Deadline: time.Now().Add(d),
		PID:      pid,
		Label:    label,
	}
	if err := state.WriteJSON(path, st); err != nil {
		if p, findErr := os.FindProcess(pid); findErr == nil {
			_ = p.Kill()
		}
		return err
	}
	return nil
}

// Stop kills the timer worker if present and removes timer.json.
func Stop(root string) error {
	path := state.Path(root, fileName)
	var st State
	err := state.ReadJSON(path, &st)
	if errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("no timer running")
	}
	if err != nil {
		return err
	}
	if p, err := os.FindProcess(st.PID); err == nil {
		_ = p.Kill()
	}
	if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return nil
}

// Status reports whether a timer file exists and how much time remains.
func Status(root string, now time.Time) (active bool, remaining time.Duration, label string, err error) {
	path := state.Path(root, fileName)
	var st State
	err = state.ReadJSON(path, &st)
	if errors.Is(err, os.ErrNotExist) {
		return false, 0, "", nil
	}
	if err != nil {
		return false, 0, "", err
	}
	remaining = st.Deadline.Sub(now)
	if remaining < 0 {
		remaining = 0
	}
	return true, remaining, st.Label, nil
}

// Worker waits until the deadline in timer.json, notifies, and clears state.
func Worker(root string, n notify.Notifier, sleep func(time.Duration), now func() time.Time) error {
	if sleep == nil {
		sleep = time.Sleep
	}
	if now == nil {
		now = time.Now
	}
	if n == nil {
		n = notify.Default{}
	}

	path := state.Path(root, fileName)
	var st State
	if err := state.ReadJSON(path, &st); err != nil {
		return err
	}

	for {
		rem := st.Deadline.Sub(now())
		if rem <= 0 {
			break
		}
		step := time.Second
		if rem < step {
			step = rem
		}
		sleep(step)
	}

	body := st.Label
	if body == "" {
		body = "Timer complete"
	}
	if err := notify.All(n, "Timer finished", body); err != nil {
		return err
	}
	if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return nil
}
