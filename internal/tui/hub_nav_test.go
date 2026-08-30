package tui

import (
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/LuizFer1/Clocky/internal/presets"
)

func TestDownArrowFocusesPresets(t *testing.T) {
	root := t.TempDir()
	_ = savePomodoroPreset(root, presets.PomodoroPreset{
		Name: "Focus", Focus: 25, Break: 5, Long: 15, Cycles: 4,
	})
	_ = savePomodoroPreset(root, presets.PomodoroPreset{
		Name: "Deep", Focus: 50, Break: 10, Long: 20, Cycles: 3,
	})
	h := newHubModel(Dependencies{Root: root, Now: time.Now})
	if h.focus != focusActions {
		t.Fatalf("want start on actions, got %v", h.focus)
	}
	m, _ := h.Update(tea.KeyMsg{Type: tea.KeyDown})
	h = m.(hubModel)
	if h.focus != focusPresets {
		t.Fatal("down should move focus to presets")
	}
	m, _ = h.Update(tea.KeyMsg{Type: tea.KeyDown})
	h = m.(hubModel)
	if h.cursor != 1 {
		t.Fatalf("second down should select next preset, cursor=%d", h.cursor)
	}
}

func TestRightArrowFocusesActions(t *testing.T) {
	h := newHubModel(Dependencies{Root: t.TempDir(), Now: time.Now})
	h.focus = focusPresets
	h.actionCursor = 0
	m, _ := h.Update(tea.KeyMsg{Type: tea.KeyRight})
	h = m.(hubModel)
	if h.focus != focusActions {
		t.Fatal("right should move focus to actions")
	}
	if h.actionCursor != 0 {
		t.Fatalf("first right from presets should only focus actions, cursor=%d", h.actionCursor)
	}
	m, _ = h.Update(tea.KeyMsg{Type: tea.KeyRight})
	h = m.(hubModel)
	if h.actionCursor != 1 {
		t.Fatalf("second right should advance action, got %d", h.actionCursor)
	}
}

func TestEnterStartsPresetWhenPresetsFocused(t *testing.T) {
	root := t.TempDir()
	_ = savePomodoroPreset(root, presets.PomodoroPreset{
		Name: "Focus", Focus: 40, Break: 10, Long: 15, Cycles: 4,
	})
	h := newHubModel(Dependencies{Root: root, Now: time.Now})
	h.focus = focusPresets
	h.cursor = 0
	_, cmd := h.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected start session cmd")
	}
	msg := cmd()
	if _, ok := msg.(startSessionMsg); !ok {
		t.Fatalf("got %#v", msg)
	}
}
