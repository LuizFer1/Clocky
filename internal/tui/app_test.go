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

func TestMinimizeKeepsSession(t *testing.T) {
	m := initialModel(Dependencies{Root: t.TempDir(), Now: time.Now})
	m.session = newSessionModel(pomodoro.Config{
		Focus: 5 * time.Minute, Break: time.Minute, Long: time.Minute, Cycles: 2, Auto: true,
	}, 80, 24, nil)
	m.page = pageSession
	before := m.session.remaining
	mod, _ := m.Update(sessionMinimizeMsg{})
	m = mod.(appModel)
	if m.page != pageHub {
		t.Fatalf("page=%v", m.page)
	}
	if !m.session.live() || m.session.remaining != before {
		t.Fatalf("session cleared or changed: live=%v rem=%v", m.session.live(), m.session.remaining)
	}
}

func TestStopPomodoroClearsSession(t *testing.T) {
	m := initialModel(Dependencies{Root: t.TempDir(), Now: time.Now})
	m.session = newSessionModel(pomodoro.Config{
		Focus: time.Minute, Break: time.Minute, Long: time.Minute, Cycles: 1, Auto: true,
	}, 80, 24, nil)
	m.page = pageHub
	m.syncHubPomodoro()
	mod, _ := m.Update(stopPomodoroMsg{})
	m = mod.(appModel)
	if m.session.live() {
		t.Fatal("expected stopped")
	}
	if m.hub.active.PomodoroActive {
		t.Fatal("active still shows pomodoro")
	}
}

func TestRefuseSecondPomodoroStart(t *testing.T) {
	m := initialModel(Dependencies{Root: t.TempDir(), Now: time.Now})
	m.session = newSessionModel(pomodoro.Config{
		Focus: time.Minute, Break: time.Minute, Long: time.Minute, Cycles: 1, Auto: true,
	}, 80, 24, nil)
	m.page = pageHub
	mod, _ := m.Update(startSessionMsg{Cfg: pomodoro.Config{
		Focus: time.Minute, Break: time.Minute, Long: time.Minute, Cycles: 1, Auto: true,
	}})
	m = mod.(appModel)
	if m.page == pageSession {
		t.Fatal("should refuse starting a second session")
	}
	if m.hub.errMsg == "" {
		t.Fatal("expected error message")
	}
}

func TestOpenPomodoroReturnsToSession(t *testing.T) {
	m := initialModel(Dependencies{Root: t.TempDir(), Now: time.Now})
	m.session = newSessionModel(pomodoro.Config{
		Focus: time.Minute, Break: time.Minute, Long: time.Minute, Cycles: 1, Auto: true,
	}, 80, 24, nil)
	m.page = pageHub
	mod, _ := m.Update(openPomodoroMsg{})
	m = mod.(appModel)
	if m.page != pageSession {
		t.Fatalf("page=%v want pageSession", m.page)
	}
}

func TestHubTickKeepsPomodoroActive(t *testing.T) {
	m := initialModel(Dependencies{Root: t.TempDir(), Now: time.Now})
	m.session = newSessionModel(pomodoro.Config{
		Focus: time.Minute, Break: time.Minute, Long: time.Minute, Cycles: 1, Auto: true,
	}, 80, 24, nil)
	m.page = pageSession
	mod, _ := m.Update(sessionMinimizeMsg{})
	m = mod.(appModel)
	m.syncHubPomodoro()
	mod, _ = m.Update(tickMsg{})
	m = mod.(appModel)
	if !m.hub.active.PomodoroActive {
		t.Fatal("expected hub.active.PomodoroActive still true after tick")
	}
}
