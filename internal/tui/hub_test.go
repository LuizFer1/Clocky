package tui

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/LuizFer1/Clocky/internal/stopwatch"
)

func TestHubViewFillsTerminal(t *testing.T) {
	h := newHubModel(Dependencies{Root: t.TempDir(), Now: time.Now})
	h.width, h.height = 100, 40
	view := h.View()
	if lipgloss.Height(view) != 40 {
		t.Fatalf("height=%d want 40", lipgloss.Height(view))
	}
}

func TestHubToggleStopwatch(t *testing.T) {
	root := t.TempDir()
	now := time.Unix(1000, 0)
	h := newHubModel(Dependencies{Root: root, Now: func() time.Time { return now }})
	m, _ := h.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'t'}})
	h = m.(hubModel)
	running, _, err := stopwatch.Status(root, now)
	if err != nil || !running {
		t.Fatalf("running=%v err=%v", running, err)
	}
	if !h.active.StopwatchRunning {
		t.Fatal("hub active not updated")
	}
}

func TestRenderActiveShowsPomodoro(t *testing.T) {
	a := activeSnapshot{
		PomodoroActive:    true,
		PomodoroPhase:     "FOCUS",
		PomodoroCycle:     1,
		PomodoroCycles:    4,
		PomodoroRemaining: 12*time.Minute + 34*time.Second,
	}
	got := renderActive(a)
	if !strings.Contains(got, "FOCUS") || !strings.Contains(got, "1/4") {
		t.Fatalf("active=%q", got)
	}
	if !strings.Contains(got, "12:34") {
		t.Fatalf("missing remaining in %q", got)
	}
}
