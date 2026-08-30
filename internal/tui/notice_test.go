package tui

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
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
	prev := h.active
	h.active = activeSnapshot{}
	if msg, ok := timerFinishedNotice(prev, h.active, h.status); ok {
		h.notice = msg
		h.status = msg
		h.alerting = true
	}
	if h.notice != "Timer finished: Break" {
		t.Fatalf("notice=%q", h.notice)
	}
	if !h.alerting {
		t.Fatal("expected alerting")
	}
	view := h.View()
	if !strings.Contains(view, "Timer finished: Break") {
		t.Fatalf("view missing notice:\n%s", view)
	}
}

func TestEscStopsAlarm(t *testing.T) {
	h := newHubModel(Dependencies{Root: t.TempDir(), Now: time.Now})
	h.notice = "Timer finished: Break"
	h.status = h.notice
	h.alerting = true
	m, _ := h.Update(tea.KeyMsg{Type: tea.KeyEsc})
	h = m.(hubModel)
	if h.alerting || h.notice != "" {
		t.Fatalf("alerting=%v notice=%q", h.alerting, h.notice)
	}
}
