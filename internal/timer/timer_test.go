package timer

import (
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/LuizFer1/Clocky/internal/state"
)

type recordingNotifier struct {
	events []string
}

func (r *recordingNotifier) Beep() error {
	r.events = append(r.events, "beep")
	return nil
}

func (r *recordingNotifier) Banner(title, body string) error {
	r.events = append(r.events, "banner:"+title+":"+body)
	return nil
}

func (r *recordingNotifier) Desktop(title, body string) error {
	r.events = append(r.events, "desktop:"+title+":"+body)
	return nil
}

func TestStatusRemaining(t *testing.T) {
	root := t.TempDir()
	now := time.Date(2026, 8, 29, 12, 0, 0, 0, time.UTC)
	deadline := now.Add(90 * time.Second)
	if err := state.EnsureDir(root); err != nil {
		t.Fatal(err)
	}
	if err := state.WriteJSON(state.Path(root, "timer.json"), State{
		Deadline: deadline,
		PID:      os.Getpid(),
		Label:    "tea",
	}); err != nil {
		t.Fatal(err)
	}

	active, remaining, label, err := Status(root, now)
	if err != nil {
		t.Fatalf("Status: %v", err)
	}
	if !active {
		t.Fatal("expected active")
	}
	if remaining != 90*time.Second {
		t.Fatalf("remaining = %v want 90s", remaining)
	}
	if label != "tea" {
		t.Fatalf("label = %q want tea", label)
	}

	// Past deadline but PID still alive → active with 0 remaining.
	active, remaining, label, err = Status(root, deadline.Add(time.Second))
	if err != nil {
		t.Fatalf("Status past: %v", err)
	}
	if !active {
		t.Fatal("expected still active while worker PID is alive")
	}
	if remaining != 0 {
		t.Fatalf("remaining = %v want 0", remaining)
	}
	if label != "tea" {
		t.Fatalf("label = %q", label)
	}
}

func TestStatusClearsStaleExpired(t *testing.T) {
	root := t.TempDir()
	if err := state.EnsureDir(root); err != nil {
		t.Fatal(err)
	}
	path := state.Path(root, "timer.json")
	if err := state.WriteJSON(path, State{
		Deadline: time.Now().Add(-time.Minute),
		PID:      1<<30 - 7, // almost certainly not running
		Label:    "stale",
	}); err != nil {
		t.Fatal(err)
	}
	active, _, _, err := Status(root, time.Now())
	if err != nil {
		t.Fatalf("Status: %v", err)
	}
	if active {
		t.Fatal("expected idle for stale expired timer")
	}
	if _, err := os.Stat(path); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("expected stale file removed, got %v", err)
	}
}

func TestStatusIdle(t *testing.T) {
	root := t.TempDir()
	active, remaining, label, err := Status(root, time.Now())
	if err != nil {
		t.Fatalf("Status: %v", err)
	}
	if active || remaining != 0 || label != "" {
		t.Fatalf("got active=%v remaining=%v label=%q", active, remaining, label)
	}
}

func TestStopWhenMissing(t *testing.T) {
	err := Stop(t.TempDir())
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "no timer") {
		t.Fatalf("error = %q want mention of no timer", err)
	}
}

func TestStopRemovesFile(t *testing.T) {
	root := t.TempDir()
	if err := state.EnsureDir(root); err != nil {
		t.Fatal(err)
	}
	path := state.Path(root, "timer.json")
	// Use a PID that is almost certainly not running.
	if err := state.WriteJSON(path, State{
		Deadline: time.Now().Add(time.Hour),
		PID:      1<<30 - 3,
		Label:    "x",
	}); err != nil {
		t.Fatal(err)
	}
	if err := Stop(root); err != nil {
		t.Fatalf("Stop: %v", err)
	}
	if _, err := os.Stat(path); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("expected file removed, got %v", err)
	}
}

func TestWorkerPastDeadlineNotifiesAndClears(t *testing.T) {
	root := t.TempDir()
	if err := state.EnsureDir(root); err != nil {
		t.Fatal(err)
	}
	deadline := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	if err := state.WriteJSON(state.Path(root, "timer.json"), State{
		Deadline: deadline,
		PID:      os.Getpid(),
		Label:    "egg",
	}); err != nil {
		t.Fatal(err)
	}

	rec := &recordingNotifier{}
	past := deadline.Add(time.Second)
	err := Worker(root, rec, func(time.Duration) {}, func() time.Time { return past })
	if err != nil {
		t.Fatalf("Worker: %v", err)
	}
	if _, err := os.Stat(state.Path(root, "timer.json")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("expected timer.json removed, got %v", err)
	}
	if len(rec.events) < 3 {
		t.Fatalf("events = %v want notify.All", rec.events)
	}
	found := false
	for _, e := range rec.events {
		if strings.HasPrefix(e, "banner:Timer finished") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("events = %v want Timer finished banner", rec.events)
	}
}

func TestStartRejectsAliveTimer(t *testing.T) {
	root := t.TempDir()
	if err := state.EnsureDir(root); err != nil {
		t.Fatal(err)
	}
	if err := state.WriteJSON(state.Path(root, "timer.json"), State{
		Deadline: time.Now().Add(time.Hour),
		PID:      os.Getpid(),
		Label:    "busy",
	}); err != nil {
		t.Fatal(err)
	}
	err := Start(root, time.Minute, "other", os.Args[0], []string{"--help"})
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "already running") {
		t.Fatalf("error = %q want already running", err)
	}
}
