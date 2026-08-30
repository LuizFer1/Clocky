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
	status       string
	notifier     notify.Notifier
}

func newSessionModel(cfg pomodoro.Config, width int, n notify.Notifier) sessionModel {
	if width <= 0 {
		width = 40
	}
	if n == nil {
		n = notify.Default{}
	}
	phases := pomodoro.PlanPhases(cfg)
	s := sessionModel{
		cfg:      cfg,
		phases:   phases,
		width:    width,
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
	if len(s.phases) == 0 {
		return "No phases configured.\n\nesc/b back  q quit"
	}
	phase := s.phases[s.index]
	face := clockface.Render(s.remaining, s.total, s.width)
	compact := clockface.RenderCompact(phase.Name, phase.Cycle, s.cfg.Cycles, s.remaining)
	barWidth := 24
	if s.width > 0 && s.width < barWidth+4 {
		barWidth = max(8, s.width-4)
	}
	bar := progressBar(s.remaining, s.total, barWidth)
	var b strings.Builder
	b.WriteString(styleTitle.Render("Pomodoro session"))
	b.WriteString("\n\n")
	b.WriteString(face)
	b.WriteString("\n")
	b.WriteString(compact)
	b.WriteString("\n")
	b.WriteString(bar)
	b.WriteString("\n")
	if s.paused {
		b.WriteString(stylePaused.Render("PAUSED"))
		b.WriteString("\n")
	}
	if s.status != "" {
		b.WriteString(styleMuted.Render(s.status))
		b.WriteString("\n")
	}
	b.WriteString("\n")
	b.WriteString(styleMuted.Render("space pause  n skip/next  esc/b hub  q quit"))
	return b.String()
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
