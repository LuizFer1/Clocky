package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/LuizFer1/Clocky/internal/clockface"
	"github.com/LuizFer1/Clocky/internal/notify"
	"github.com/LuizFer1/Clocky/internal/pomodoro"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type sessionDoneMsg struct{}
type sessionAbortMsg struct{}

type sessionMinimizeMsg struct{}

type sessionPhaseEndMsg struct {
	Title string
	Body  string
	Done  bool // true when the whole pomodoro finished
}

type sessionModel struct {
	cfg          pomodoro.Config
	phases       []pomodoro.Phase
	index        int
	remaining    time.Duration
	total        time.Duration
	paused       bool
	waitingEnter bool
	active       bool
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
		s.active = true
		s.total = phases[0].Duration
		s.remaining = phases[0].Duration
		if s.total <= 0 {
			s.total = time.Second
			s.remaining = time.Second
		}
	}
	return s
}

func (s sessionModel) live() bool { return s.active }

func (s sessionModel) Init() tea.Cmd {
	return nil // the app model owns the single tick loop
}

func (s sessionModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		s.width = msg.Width
		s.height = msg.Height
		return s, nil
	case tickMsg:
		if s.waitingEnter || s.paused || len(s.phases) == 0 {
			return s, nil
		}
		step := time.Second
		if s.remaining < step {
			step = s.remaining
		}
		s.remaining -= step
		if s.remaining <= 0 {
			return s.finishPhase()
		}
		return s, nil
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
			return s, func() tea.Msg { return sessionMinimizeMsg{} }
		case "q", "ctrl+c":
			return s, tea.Quit
		}
	}
	return s, nil
}

func (s sessionModel) finishPhase() (tea.Model, tea.Cmd) {
	phase := s.phases[s.index]
	title, body := phaseEndTitleBody(phase)
	done := s.index >= len(s.phases)-1
	phaseCmd := func() tea.Msg {
		return sessionPhaseEndMsg{Title: title, Body: body, Done: done}
	}
	if done {
		s.status = "Session complete"
		s.remaining = 0
		return s, tea.Batch(phaseCmd, func() tea.Msg { return sessionDoneMsg{} })
	}
	if !s.cfg.Auto {
		s.waitingEnter = true
		s.paused = false
		s.status = "Press Enter or n to continue"
		return s, phaseCmd
	}
	// Auto: advance, then notify about the phase that just ended.
	m, advCmd := s.advanceAfterWait()
	return m, tea.Batch(phaseCmd, advCmd)
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
	return s, nil
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
	face := colorFace(clockface.Render(s.remaining, s.total, faceW), phase.Name)
	compact := phaseLine(phase.Name, phase.Cycle, s.cfg.Cycles, s.remaining)
	barWidth := min(36, max(16, faceW-4))
	bar := progressBar(s.remaining, s.total, barWidth)

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

// phaseColor picks the accent for a pomodoro phase.
func phaseColor(name string) lipgloss.Color {
	switch name {
	case "FOCUS":
		return colAccent2
	case "BREAK":
		return colOK
	default:
		return colWarn
	}
}

// phaseLine renders "● FOCUS  1/4  24:48" with the phase accent.
func phaseLine(name string, cycle, cycles int, remaining time.Duration) string {
	acc := lipgloss.NewStyle().Bold(true).Foreground(phaseColor(name))
	timeStyle := lipgloss.NewStyle().Bold(true).Foreground(colText)
	sec := int64(remaining / time.Second)
	return acc.Render("● "+name) +
		styleMuted.Render(fmt.Sprintf("  %d/%d  ", cycle, cycles)) +
		timeStyle.Render(fmt.Sprintf("%02d:%02d", sec/60, sec%60))
}

// colorFace applies colours to the plain clockface by glyph class.
func colorFace(face, phase string) string {
	num := lipgloss.NewStyle().Bold(true).Foreground(colAccent)
	minute := lipgloss.NewStyle().Bold(true).Foreground(phaseColor(phase))
	center := lipgloss.NewStyle().Bold(true).Foreground(colText)

	var b strings.Builder
	for _, r := range face {
		switch {
		case r == '\n' || r == ' ':
			b.WriteRune(r)
		case r == clockface.RimRune:
			b.WriteString(styleDim.Render(string(r)))
		case r == clockface.TickRune:
			b.WriteString(styleMuted.Render(string(r)))
		case r == clockface.CenterRune:
			b.WriteString(center.Render(string(r)))
		case r >= '0' && r <= '9':
			b.WriteString(num.Render(string(r)))
		case strings.ContainsRune(clockface.MinuteHandRunes, r):
			b.WriteString(minute.Render(string(r)))
		case strings.ContainsRune(clockface.SecondHandRunes, r):
			b.WriteString(styleMuted.Render(string(r)))
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
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
	fill := lipgloss.NewStyle().Foreground(colAccent).Render(strings.Repeat("━", filled))
	return fill + styleDim.Render(strings.Repeat("╌", width-filled))
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
