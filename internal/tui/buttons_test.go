package tui

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

func TestHubActionBarInView(t *testing.T) {
	h := newHubModel(Dependencies{Root: t.TempDir(), Now: time.Now})
	h.width, h.height = 100, 40
	view := h.View()
	for _, label := range []string{"New Pomodoro", "New Timer", "Start", "Stop", "Stopwatch"} {
		if !strings.Contains(view, label) {
			t.Fatalf("missing button %q in:\n%s", label, view)
		}
	}
}

func TestActivateNewPomodoroOpensLaunchForm(t *testing.T) {
	h := newHubModel(Dependencies{Root: t.TempDir(), Now: time.Now})
	h.focus = focusActions
	h.actionCursor = 0 // New Pomodoro
	_, cmd := h.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected command")
	}
	msg := cmd()
	launch, ok := msg.(openLaunchFormMsg)
	if !ok || launch.Kind != formPomodoro {
		t.Fatalf("got %#v", msg)
	}
}

func TestLaunchPomodoroFormSubmit(t *testing.T) {
	root := t.TempDir()
	f := newLaunchPomodoroForm(root)
	f.fields = []string{"25", "5", "15", "2", "y", "Deep"}
	cmd, err := f.submitLaunch()
	if err != nil {
		t.Fatal(err)
	}
	msg := cmd().(formLaunchPomodoroMsg)
	if !msg.SavedPreset || msg.SavedName != "Deep" {
		t.Fatalf("%+v", msg)
	}
	if msg.Cfg.Cycles != 2 || !msg.Cfg.Auto {
		t.Fatalf("cfg=%+v", msg.Cfg)
	}
}
