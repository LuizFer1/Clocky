package tui

import (
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/LuizFer1/Clocky/internal/stopwatch"
)

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
