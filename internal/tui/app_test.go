package tui

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/LuizFer1/Clocky/internal/pomodoro"
)

func pomodoroConfigForTest() pomodoro.Config {
	return pomodoro.Config{
		Focus: 5 * time.Second, Break: time.Second, Long: time.Second, Cycles: 1, Auto: true,
	}
}

func lipglossHeight(s string) int { return lipgloss.Height(s) }

func TestHubViewHasActiveAndPresets(t *testing.T) {
	m := initialModel(Dependencies{Root: t.TempDir(), Now: time.Now})
	m.width, m.height = 80, 24
	m.hub.width, m.hub.height = 80, 24
	view := m.View()
	if !strings.Contains(view, "Active") || !strings.Contains(view, "Presets") {
		t.Fatalf("view missing panels:\n%s", view)
	}
}

func TestSessionViewFillsAndCenters(t *testing.T) {
	cfg := pomodoroConfigForTest()
	s := newSessionModel(cfg, 80, 30, nil)
	view := s.View()
	if lipglossHeight(view) != 30 {
		t.Fatalf("height=%d want 30", lipglossHeight(view))
	}
	if !strings.Contains(view, "Pomodoro") {
		t.Fatalf("missing header:\n%s", view)
	}
}

func TestHubQuitCommand(t *testing.T) {
	m := initialModel(Dependencies{Root: t.TempDir(), Now: time.Now})
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	if cmd == nil {
		t.Fatal("expected quit command")
	}
}
