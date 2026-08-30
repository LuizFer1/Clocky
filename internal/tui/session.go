package tui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/LuizFer1/Clocky/internal/clockface"
	"github.com/LuizFer1/Clocky/internal/notify"
	"github.com/LuizFer1/Clocky/internal/pomodoro"
)

type sessionDoneMsg struct{}
type sessionAbortMsg struct{}

type sessionModel struct {
	cfg          pomodoro.Config
	phases       []pomodoro.Phase
	index        int
	remaining    time.Duration
	total        time.Duration
	paused       bool
	waitingEnter bool
	width        int
	height       int
	status       string
	notifier     notify.Notifier
}

func newSessionModel(cfg pomodoro.Config, width, height int, n notify.Notifier) sessionModel {
	if width <= 0 {
		width = 80
	}
	if height <= 0 {
		height = 24
	}
	if n == nil {
		n = notify.Default{}
	}
	phases := pomodoro.PlanPhases(cfg)
	s := sessionModel{
		cfg:      cfg,
		phases:   phases,
		width:    width,
		height:   height,
		notifier: n,
	}
	if len(phases) > 0 {
		s.total = phases[0].Duration
		s.remaining = phases[0].Duration
		if s.total <= 0 {
			s.total = time.Second
			s.remaining = time.Second
		}
	}
	return s
}

func (s sessionModel) Init() tea.Cmd {
	return scheduleTick()
}

func (s sessionModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		s.width = msg.Width
		s.height = msg.Height
		return s, nil
	case tickMsg:
		if s.waitingEnter || s.paused || len(s.phases) == 0 {
			return s, scheduleTick()
		}
		step := time.Second
		if s.remaining < step {
			step = s.remaining
		}
		s.remaining -= step
		if s.remaining <= 0 {
			return s.finishPhase()
		}
		return s, scheduleTick()
	case tea.KeyMsg:
		switch msg.String() {
		case " ":
			if !s.waitingEnter {
				s.paused = !s.paused
				if s.paused {
					s.status = "Paused"
				} else {
					s.status = ""
				}
			}
			return s, nil
		case "n":
			if s.waitingEnter {
				return s.advanceAfterWait()
			}
			s.remaining = 0
			return s.finishPhase()
		case "enter":
			if s.waitingEnter {
				return s.advanceAfterWait()
			}
			return s, nil
		case "esc", "b":
			return s, func() tea.Msg { return sessionAbortMsg{} }
		case "q", "ctrl+c":
			return s, tea.Quit
		}
	}
	return s, nil
}

func (s sessionModel) finishPhase() (tea.Model, tea.Cmd) {
	phase := s.phases[s.index]
	title, body := phaseEndTitleBody(phase)
	_ = notify.All(s.notifier, title, body)
	if s.index >= len(s.phases)-1 {
		s.status = "Session complete"
		return s, func() tea.Msg { return sessionDoneMsg{} }
	}
	if !s.cfg.Auto {
		s.waitingEnter = true
		s.paused = false
		s.status = "Press Enter or n to continue"
		return s, scheduleTick()
	}
	return s.advanceAfterWait()
}

func (s sessionModel) advanceAfterWait() (tea.Model, tea.Cmd) {
	s.waitingEnter = false
	s.status = ""
	s.index++
	if s.index >= len(s.phases) {
		return s, func() tea.Msg { return sessionDoneMsg{} }
	}
	ph := s.phases[s.index]
	s.total = ph.Duration
	s.remaining = ph.Duration
	if s.total <= 0 {
		s.total = time.Second
		s.remaining = time.Second
	}
	s.paused = false
	return s, scheduleTick()
}

func (s sessionModel) View() string {
	w, ht := s.width, s.height
	if w <= 0 {
		w = 80
	}
	if ht <= 0 {
		ht = 24
	}
	footer := "space pause  n skip/next  esc/b hub  q quit"
	if len(s.phases) == 0 {
		return fillFrame(w, ht, "Pomodoro", styleMuted.Render("No phases configured."), footer)
	}
	phase := s.phases[s.index]
	// Size the ASCII face for the available width (leave room for frame chrome).
	faceW := w - 6
	if faceW < 15 {
		faceW = w
	}
	if faceW > 51 {
		faceW = 51
	}
	face := clockface.Render(s.remaining, s.total, faceW)
	compact := stylePhase.Render(clockface.RenderCompact(phase.Name, phase.Cycle, s.cfg.Cycles, s.remaining))
	barWidth := min(36, max(16, faceW-4))
	bar := styleTitle.Render(progressBar(s.remaining, s.total, barWidth))

	parts := []string{face, "", compact, bar}
	if s.paused {
		parts = append(parts, "", stylePaused.Render("◆ PAUSED ◆"))
	}
	if s.status != "" {
		parts = append(parts, "", styleMuted.Render(s.status))
	}
	body := lipglossJoin(parts...)
	return fillFrame(w, ht, "Pomodoro", body, footer)
}

func lipglossJoin(parts ...string) string {
	return strings.Join(parts, "\n")
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func progressBar(remaining, total time.Duration, width int) string {
	if width <= 0 {
		return ""
	}
	if total <= 0 {
		return strings.Repeat("░", width)
	}
	done := 1 - float64(remaining)/float64(total)
	if done < 0 {
		done = 0
	}
	if done > 1 {
		done = 1
	}
	filled := int(done * float64(width))
	if filled > width {
		filled = width
	}
	return strings.Repeat("█", filled) + strings.Repeat("░", width-filled)
}

func phaseEndTitleBody(phase pomodoro.Phase) (title, body string) {
	switch phase.Name {
	case "FOCUS":
		return "Focus complete", fmt.Sprintf("Focus session %d finished", phase.Cycle)
	case "BREAK":
		return "Break complete", fmt.Sprintf("Short break after focus %d finished", phase.Cycle)
	default:
		return "Long break complete", fmt.Sprintf("Long break after focus %d finished", phase.Cycle)
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
