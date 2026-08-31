package tui

import (
	"strings"
	"testing"
	"time"

	"github.com/LuizFer1/Clocky/internal/notify"
	"github.com/LuizFer1/Clocky/internal/pomodoro"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func TestSessionPauseFreezesRemaining(t *testing.T) {
	cfg := pomodoro.Config{
		Focus:  5 * time.Second,
		Break:  time.Second,
		Long:   time.Second,
		Cycles: 1,
		Auto:   true,
	}
	s := newSessionModel(cfg, 40, 24, nil)
	m, _ := s.Update(tickMsg{})
	s = m.(sessionModel)
	if s.remaining != 4*time.Second {
		t.Fatalf("after tick remaining=%v", s.remaining)
	}
	m, _ = s.Update(tea.KeyMsg{Type: tea.KeySpace})
	s = m.(sessionModel)
	if !s.paused {
		t.Fatal("expected paused")
	}
	before := s.remaining
	m, _ = s.Update(tickMsg{})
	s = m.(sessionModel)
	if s.remaining != before {
		t.Fatalf("paused remaining changed: %v -> %v", before, s.remaining)
	}
}

func TestProgressBarBounds(t *testing.T) {
	bar := progressBar(5*time.Second, 10*time.Second, 10)
	if lipgloss.Width(bar) != 10 {
		t.Fatalf("bar=%q len=%d", bar, lipgloss.Width(bar))
	}
}

func TestFinishPhaseDoesNotCallBanner(t *testing.T) {
	rec := &notify.RecordingNotifier{}
	cfg := pomodoro.Config{
		Focus:  time.Second,
		Break:  time.Second,
		Long:   time.Second,
		Cycles: 1,
		Auto:   true,
	}
	s := newSessionModel(cfg, 40, 24, rec)
	s.remaining = 0
	m, cmd := s.finishPhase()
	s = m.(sessionModel)
	_ = s
	if cmd == nil {
		t.Fatal("expected command from finishPhase")
	}
	end, ok := findSessionPhaseEndMsg(cmd)
	if !ok {
		t.Fatal("expected sessionPhaseEndMsg from finishPhase")
	}
	if end.Title != "Focus complete" {
		t.Fatalf("title=%q", end.Title)
	}
	for _, e := range rec.Events {
		if strings.HasPrefix(e, "banner:") {
			t.Fatalf("banner must not be called from TUI finishPhase: %v", rec.Events)
		}
	}
}

func findSessionPhaseEndMsg(cmd tea.Cmd) (sessionPhaseEndMsg, bool) {
	if cmd == nil {
		return sessionPhaseEndMsg{}, false
	}
	msg := cmd()
	if end, ok := msg.(sessionPhaseEndMsg); ok {
		return end, true
	}
	if batch, ok := msg.(tea.BatchMsg); ok {
		for _, c := range batch {
			if end, found := findSessionPhaseEndMsg(c); found {
				return end, true
			}
		}
	}
	return sessionPhaseEndMsg{}, false
}
