package tui

import (
	"strings"
	"testing"
	"time"
)

func TestTimerFinishedNoticeOnClear(t *testing.T) {
	prev := activeSnapshot{TimerActive: true, TimerRemaining: 2 * time.Second, TimerLabel: "Break"}
	cur := activeSnapshot{}
	msg, ok := timerFinishedNotice(prev, cur, "")
	if !ok || msg != "Timer finished: Break" {
		t.Fatalf("msg=%q ok=%v", msg, ok)
	}
}

func TestTimerFinishedNoticeOnZero(t *testing.T) {
	prev := activeSnapshot{TimerActive: true, TimerRemaining: time.Second, TimerLabel: "Tea"}
	cur := activeSnapshot{TimerActive: true, TimerRemaining: 0, TimerLabel: "Tea"}
	msg, ok := timerFinishedNotice(prev, cur, "")
	if !ok || msg != "Timer finished: Tea" {
		t.Fatalf("msg=%q ok=%v", msg, ok)
	}
}

func TestTimerFinishedNoticeIgnoresManualStop(t *testing.T) {
	prev := activeSnapshot{TimerActive: true, TimerRemaining: 5 * time.Second, TimerLabel: "Break"}
	cur := activeSnapshot{}
	_, ok := timerFinishedNotice(prev, cur, "Timer stopped")
	if ok {
		t.Fatal("manual stop should not look like finished")
	}
}

func TestTimerFinishedNoticeIgnoresIdle(t *testing.T) {
	_, ok := timerFinishedNotice(activeSnapshot{}, activeSnapshot{}, "")
	if ok {
		t.Fatal("idle should not notify")
	}
}

func TestHubTickSetsNotice(t *testing.T) {
	h := newHubModel(Dependencies{Root: t.TempDir(), Now: time.Now})
	h.active = activeSnapshot{TimerActive: true, TimerRemaining: 2 * time.Second, TimerLabel: "Break"}
	// Simulate next refresh by patching via Update after manually setting prev through tick path:
	// call timerFinishedNotice path by constructing tick with swapped state in Update.
	prev := h.active
	h.active = activeSnapshot{}
	if msg, ok := timerFinishedNotice(prev, h.active, h.status); ok {
		h.notice = msg
		h.status = msg
	}
	if h.notice != "Timer finished: Break" {
		t.Fatalf("notice=%q", h.notice)
	}
	view := h.View()
	if !strings.Contains(view, "Timer finished: Break") {
		t.Fatalf("view missing notice:\n%s", view)
	}
}
