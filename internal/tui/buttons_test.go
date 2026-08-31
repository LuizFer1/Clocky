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
	for _, label := range []string{"New Pomodoro", "New Timer", "Start Stopwatch"} {
		if !strings.Contains(view, label) {
			t.Fatalf("missing button %q in:\n%s", label, view)
		}
	}
	if strings.Contains(view, "[ Start ]") || strings.Contains(view, "[ Stop ]") {
		t.Fatal("unexpected standalone Start/Stop buttons")
	}
}

func TestStopwatchButtonLabelFlips(t *testing.T) {
	root := t.TempDir()
	now := time.Unix(1000, 0)
	h := newHubModel(Dependencies{Root: root, Now: func() time.Time { return now }})
	btns := hubActionButtons(h.active)
	if btns[2].Label != "Start Stopwatch" {
		t.Fatalf("idle label=%q", btns[2].Label)
	}
	m, _ := h.doStopwatch()
	h = m.(hubModel)
	btns = hubActionButtons(h.active)
	if btns[2].Label != "Stop Stopwatch" {
		t.Fatalf("running label=%q", btns[2].Label)
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

func TestHubActionButtonsIncludePomodoroControls(t *testing.T) {
	a := activeSnapshot{PomodoroActive: true}
	btns := hubActionButtons(a)
	labels := make([]string, len(btns))
	for i, b := range btns {
		labels[i] = b.Label
	}
	for _, label := range labels {
		if label == "New Pomodoro" {
			t.Fatalf("New Pomodoro should be absent when active; buttons=%v", labels)
		}
	}
	if len(btns) < 2 || btns[0].Label != "Open Pomodoro" || btns[1].Label != "Stop Pomodoro" {
		t.Fatalf("expected Open then Stop first; buttons=%v", labels)
	}
}

func TestActivateOpenAndStopPomodoro(t *testing.T) {
	h := newHubModel(Dependencies{Root: t.TempDir(), Now: time.Now})

	_, cmd := h.activateAction(actionOpenPomodoro)
	if cmd == nil {
		t.Fatal("expected open command")
	}
	msg := cmd()
	if _, ok := msg.(openPomodoroMsg); !ok {
		t.Fatalf("expected openPomodoroMsg, got %#v", msg)
	}

	_, cmd = h.activateAction(actionStopPomodoro)
	if cmd == nil {
		t.Fatal("expected stop command")
	}
	msg = cmd()
	if _, ok := msg.(stopPomodoroMsg); !ok {
		t.Fatalf("expected stopPomodoroMsg, got %#v", msg)
	}
}
