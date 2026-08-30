package pomodoro

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/LuizFer1/Clocky/internal/notify"
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

func TestPlanPhasesCycles2(t *testing.T) {
	cfg := Config{
		Focus:  25 * time.Minute,
		Break:  5 * time.Minute,
		Long:   15 * time.Minute,
		Cycles: 2,
	}
	phases := PlanPhases(cfg)
	want := []Phase{
		{Name: "FOCUS", Duration: 25 * time.Minute, Cycle: 1},
		{Name: "BREAK", Duration: 5 * time.Minute, Cycle: 1},
		{Name: "FOCUS", Duration: 25 * time.Minute, Cycle: 2},
		{Name: "LONG BREAK", Duration: 15 * time.Minute, Cycle: 2},
	}
	assertPhases(t, phases, want)
}

func TestPlanPhasesCycles4(t *testing.T) {
	cfg := Config{
		Focus:  time.Minute,
		Break:  30 * time.Second,
		Long:   2 * time.Minute,
		Cycles: 4,
	}
	phases := PlanPhases(cfg)
	want := []Phase{
		{Name: "FOCUS", Duration: time.Minute, Cycle: 1},
		{Name: "BREAK", Duration: 30 * time.Second, Cycle: 1},
		{Name: "FOCUS", Duration: time.Minute, Cycle: 2},
		{Name: "BREAK", Duration: 30 * time.Second, Cycle: 2},
		{Name: "FOCUS", Duration: time.Minute, Cycle: 3},
		{Name: "BREAK", Duration: 30 * time.Second, Cycle: 3},
		{Name: "FOCUS", Duration: time.Minute, Cycle: 4},
		{Name: "LONG BREAK", Duration: 2 * time.Minute, Cycle: 4},
	}
	assertPhases(t, phases, want)
}

func TestPlanPhasesCycles1(t *testing.T) {
	cfg := Config{
		Focus:  10 * time.Second,
		Break:  5 * time.Second,
		Long:   20 * time.Second,
		Cycles: 1,
	}
	phases := PlanPhases(cfg)
	want := []Phase{
		{Name: "FOCUS", Duration: 10 * time.Second, Cycle: 1},
		{Name: "LONG BREAK", Duration: 20 * time.Second, Cycle: 1},
	}
	assertPhases(t, phases, want)
}

func TestRunAutoNotifiesEachPhase(t *testing.T) {
	rec := &recordingNotifier{}
	var out bytes.Buffer
	cfg := Config{
		Focus:  time.Millisecond,
		Break:  time.Millisecond,
		Long:   time.Millisecond,
		Cycles: 1,
		Auto:   true,
	}
	err := Run(cfg, Hooks{
		Notify: rec,
		Sleep:  func(time.Duration) {},
		Now:    time.Now,
		In:     strings.NewReader(""),
		Out:    &out,
		Width:  21,
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	// Each phase end: beep + banner + desktop → 2 phases × 3 = 6 events.
	if len(rec.events) < 6 {
		t.Fatalf("events = %v want at least 6 (2 phase ends)", rec.events)
	}
	beeps := 0
	for _, e := range rec.events {
		if e == "beep" {
			beeps++
		}
	}
	if beeps < 2 {
		t.Fatalf("beeps = %d want >= 2; events=%v", beeps, rec.events)
	}
	if out.Len() == 0 {
		t.Fatal("expected clockface output")
	}
}

func TestRunWaitsForEnterWhenNotAuto(t *testing.T) {
	rec := &recordingNotifier{}
	cfg := Config{
		Focus:  time.Millisecond,
		Break:  time.Millisecond,
		Long:   time.Millisecond,
		Cycles: 1,
		Auto:   false,
	}
	err := Run(cfg, Hooks{
		Notify: rec,
		Sleep:  func(time.Duration) {},
		In:     strings.NewReader("\n"),
		Out:    &bytes.Buffer{},
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(rec.events) < 6 {
		t.Fatalf("events = %v want at least 6", rec.events)
	}
}

func assertPhases(t *testing.T, got, want []Phase) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("len = %d want %d\ngot=%v", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("phase[%d] = %+v want %+v", i, got[i], want[i])
		}
	}
}

// Ensure notify.All path compiles against the interface.
var _ notify.Notifier = (*recordingNotifier)(nil)
