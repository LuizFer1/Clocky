package tui

import (
	"testing"
	"time"

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
