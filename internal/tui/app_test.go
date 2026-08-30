package tui

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

func TestHubViewHasActiveAndPresets(t *testing.T) {
	m := initialModel(Dependencies{Root: t.TempDir(), Now: time.Now})
	view := m.View()
	if !strings.Contains(view, "Active") || !strings.Contains(view, "Presets") {
		t.Fatalf("view missing panels:\n%s", view)
	}
}

func TestHubQuitCommand(t *testing.T) {
	m := initialModel(Dependencies{Root: t.TempDir(), Now: time.Now})
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	if cmd == nil {
		t.Fatal("expected quit command")
	}
}
