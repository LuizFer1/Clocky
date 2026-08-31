package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/LuizFer1/Clocky/internal/duration"
	"github.com/LuizFer1/Clocky/internal/notify"
	"github.com/LuizFer1/Clocky/internal/pomodoro"
	"github.com/LuizFer1/Clocky/internal/presets"
	"github.com/LuizFer1/Clocky/internal/stopwatch"
	"github.com/LuizFer1/Clocky/internal/timer"
	tea "github.com/charmbracelet/bubbletea"
)

type hubAlertMsg struct{}

func hubAudioAlert() tea.Cmd {
	return func() tea.Msg {
		notify.Alert()
		return hubAlertMsg{}
	}
}

type startSessionMsg struct {
	Cfg pomodoro.Config
}

type openPickerMsg struct{}

type openFormMsg struct {
	Kind formKind
	Pomo *presets.PomodoroPreset
	Tim  *presets.TimerPreset
}

type openLaunchFormMsg struct {
	Kind formKind
}

type openConfirmMsg struct {
	Kind string
	Name string
}

type openPomodoroMsg struct{}
type stopPomodoroMsg struct{}

type hubModel struct {
	deps         Dependencies
	active       activeSnapshot
	items        []presetItem
	cursor       int
	actionCursor int
	focus        hubFocus
	status       string
	errMsg       string
	notice       string // prominent in-hub alert (e.g. timer finished)
	alerting     bool   // keep sounding until the user dismisses the notice
	width        int
	height       int
}

func newHubModel(deps Dependencies) hubModel {
	h := hubModel{deps: deps, width: 80, height: 24, focus: focusActions}
	h.reload()
	return h
}

func (h *hubModel) reload() {
	now := h.deps.Now()
	h.active = refreshActive(h.deps.Root, now)
	items, err := loadPresetItems(h.deps.Root)
	if err != nil {
		h.errMsg = err.Error()
		h.items = nil
		return
	}
	h.items = items
	if len(h.items) == 0 {
		h.cursor = 0
		return
	}
	if h.cursor >= len(h.items) {
		h.cursor = len(h.items) - 1
	}
}

func (h hubModel) Init() tea.Cmd {
	return scheduleTick()
}

func (h hubModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		h.width = msg.Width
		h.height = msg.Height
		return h, nil
	case tickMsg:
		prev := h.active
		h.active = refreshActive(h.deps.Root, h.deps.Now())
		if note, ok := timerFinishedNotice(prev, h.active, h.status); ok {
			h.notice = note
			h.status = note
			h.errMsg = ""
			h.alerting = true
			return h, tea.Batch(scheduleTick(), hubAudioAlert(), scheduleAlertTick())
		}
		return h, scheduleTick()
	case alertTickMsg:
		if !h.alerting {
			return h, nil
		}
		return h, tea.Batch(hubAudioAlert(), scheduleAlertTick())
	case hubAlertMsg:
		return h, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			h.alerting = false
			notify.StopAlert()
			return h, tea.Quit
		case "esc":
			if h.notice != "" || h.alerting {
				was := h.notice
				h.notice = ""
				h.alerting = false
				notify.StopAlert()
				if h.status == was || strings.HasPrefix(h.status, "Timer finished:") {
					h.status = ""
				}
				return h, nil
			}
			return h, nil
		case "tab":
			if h.focus == focusActions {
				h.focus = focusPresets
			} else {
				h.focus = focusActions
			}
			return h, nil
		case "left", "h":
			// Horizontal keys drive the action bar.
			if h.focus != focusActions {
				h.focus = focusActions
				return h, nil
			}
			if h.actionCursor > 0 {
				h.actionCursor--
			}
			return h, nil
		case "right", "l":
			if h.focus != focusActions {
				h.focus = focusActions
				return h, nil
			}
			btns := hubActionButtons(h.active)
			if h.actionCursor < len(btns)-1 {
				h.actionCursor++
			}
			return h, nil
		case "up", "k":
			// Vertical keys drive the preset list.
			if h.focus != focusPresets {
				h.focus = focusPresets
				return h, nil
			}
			if h.cursor > 0 {
				h.cursor--
			}
			return h, nil
		case "down", "j":
			if h.focus != focusPresets {
				h.focus = focusPresets
				return h, nil
			}
			if h.cursor < len(h.items)-1 {
				h.cursor++
			}
			return h, nil
		case "enter":
			if h.focus == focusActions {
				btns := hubActionButtons(h.active)
				if len(btns) == 0 {
					return h, nil
				}
				if h.actionCursor >= len(btns) {
					h.actionCursor = len(btns) - 1
				}
				return h.activateAction(btns[h.actionCursor].ID)
			}
			return h.startSelected()
		case " ":
			// Space always activates the highlighted action button.
			h.focus = focusActions
			btns := hubActionButtons(h.active)
			if len(btns) == 0 {
				return h, nil
			}
			if h.actionCursor >= len(btns) {
				h.actionCursor = len(btns) - 1
			}
			return h.activateAction(btns[h.actionCursor].ID)
		case "s":
			if h.active.TimerActive {
				return h.doStopTimer()
			}
			return h.doStopwatch()
		case "t":
			return h.doStopwatch()
		case "p":
			return h, func() tea.Msg { return openLaunchFormMsg{Kind: formPomodoro} }
		case "n":
			// keep picker for generic "new preset", but prefer launch buttons
			return h, func() tea.Msg { return openPickerMsg{} }
		case "e":
			if len(h.items) == 0 {
				h.errMsg = "no preset selected"
				return h, nil
			}
			it := h.items[h.cursor]
			if it.Kind == kindPomodoro && it.Pomo != nil {
				p := *it.Pomo
				return h, func() tea.Msg { return openFormMsg{Kind: formPomodoro, Pomo: &p} }
			}
			if it.Kind == kindTimer && it.Tim != nil {
				tm := *it.Tim
				return h, func() tea.Msg { return openFormMsg{Kind: formTimer, Tim: &tm} }
			}
			return h, nil
		case "d":
			if len(h.items) == 0 {
				h.errMsg = "no preset selected"
				return h, nil
			}
			it := h.items[h.cursor]
			return h, func() tea.Msg { return openConfirmMsg{Kind: it.Kind, Name: it.Name} }
		}
	}
	return h, nil
}

func (h hubModel) activateAction(id actionID) (tea.Model, tea.Cmd) {
	switch id {
	case actionNewPomodoro:
		return h, func() tea.Msg { return openLaunchFormMsg{Kind: formPomodoro} }
	case actionNewTimer:
		return h, func() tea.Msg { return openLaunchFormMsg{Kind: formTimer} }
	case actionStopwatchToggle:
		return h.doStopwatch()
	case actionStopTimer:
		return h.doStopTimer()
	case actionOpenPomodoro:
		return h, func() tea.Msg { return openPomodoroMsg{} }
	case actionStopPomodoro:
		return h, func() tea.Msg { return stopPomodoroMsg{} }
	default:
		return h, nil
	}
}

func (h hubModel) doStopTimer() (tea.Model, tea.Cmd) {
	if err := timer.Stop(h.deps.Root); err != nil {
		h.errMsg = err.Error()
	} else {
		h.status = "Timer stopped"
		h.errMsg = ""
		h.notice = ""
		h.alerting = false
		notify.StopAlert()
	}
	h.active = refreshActive(h.deps.Root, h.deps.Now())
	btns := hubActionButtons(h.active)
	if h.actionCursor >= len(btns) && len(btns) > 0 {
		h.actionCursor = len(btns) - 1
	}
	return h, nil
}

func (h hubModel) doStopwatch() (tea.Model, tea.Cmd) {
	action, elapsed, err := stopwatch.Toggle(h.deps.Root, h.deps.Now())
	if err != nil {
		h.errMsg = err.Error()
	} else if action == "started" {
		h.status = "Stopwatch started"
		h.errMsg = ""
	} else {
		h.status = fmt.Sprintf("Stopwatch stopped: %s", duration.Format(elapsed))
		h.errMsg = ""
	}
	h.active = refreshActive(h.deps.Root, h.deps.Now())
	return h, nil
}

func (h hubModel) startSelected() (tea.Model, tea.Cmd) {
	if len(h.items) == 0 {
		h.errMsg = "no preset selected — use New Pomodoro / New Timer, or Enter on a preset"
		return h, nil
	}
	it := h.items[h.cursor]
	switch it.Kind {
	case kindPomodoro:
		if it.Pomo == nil {
			return h, nil
		}
		p := it.Pomo
		return h, func() tea.Msg {
			return startSessionMsg{Cfg: pomodoro.Config{
				Focus:  time.Duration(p.Focus) * time.Minute,
				Break:  time.Duration(p.Break) * time.Minute,
				Long:   time.Duration(p.Long) * time.Minute,
				Cycles: p.Cycles,
				Auto:   p.Auto,
			}}
		}
	case kindTimer:
		if it.Tim == nil {
			return h, nil
		}
		return h.startTimer(time.Duration(it.Tim.Seconds)*time.Second, it.Name)
	}
	return h, nil
}

func (h hubModel) startTimer(d time.Duration, label string) (tea.Model, tea.Cmd) {
	exe, err := h.deps.Exe()
	if err != nil {
		h.errMsg = err.Error()
		return h, nil
	}
	if err := timer.Start(h.deps.Root, d, label, exe, []string{"timer", "--worker"}); err != nil {
		h.errMsg = err.Error()
		return h, nil
	}
	h.status = fmt.Sprintf("Timer started: %s (%s)", label, duration.Format(d))
	h.errMsg = ""
	h.active = refreshActive(h.deps.Root, h.deps.Now())
	return h, nil
}

func (h hubModel) View() string {
	w, ht := h.width, h.height
	if w <= 0 {
		w = 80
	}
	if ht <= 0 {
		ht = 24
	}
	panelW := w - 8
	if panelW > 72 {
		panelW = 72
	}
	if panelW < 28 {
		panelW = max(20, w-4)
	}

	btns := hubActionButtons(h.active)
	if h.actionCursor >= len(btns) && len(btns) > 0 {
		h.actionCursor = len(btns) - 1
	}
	actions := wrapActionBar(btns, h.actionCursor, h.focus == focusActions, panelW, w)
	activePanel := panelBox("Active", renderActive(h.active), panelW)
	presetsPanel := panelBoxFocused("Presets", renderPresets(h.items, h.cursor, h.focus == focusPresets), panelW, h.focus == focusPresets)
	noticePanel := renderNotice(h.notice, panelW)

	var statusLine string
	if h.errMsg != "" {
		statusLine = styleError.Render(h.errMsg)
	} else if h.status != "" && h.status != h.notice {
		statusLine = styleOK.Render(h.status)
	}

	body := joinPanels(noticePanel, actions, activePanel, presetsPanel, statusLine)
	footer := "←→ actions  ↑↓ presets  enter  space action  esc stop alarm  e edit  d delete  q quit"
	return fillFrame(w, ht, "hub", body, footer)
}

func renderActive(a activeSnapshot) string {
	on := styleOK.Render("●")
	off := styleDim.Render("○")
	name := func(s string) string { return fmt.Sprintf("%-10s", s) }
	var lines []string
	if a.TimerActive {
		label := a.TimerLabel
		if label == "" {
			label = "timer"
		}
		if a.TimerRemaining == 0 {
			lines = append(lines, on+" "+name("Timer")+" "+stylePaused.Render("FINISHED")+"  "+styleMuted.Render(label))
		} else {
			lines = append(lines, on+" "+name("Timer")+" "+styleOK.Render(duration.Format(a.TimerRemaining)+" left")+"  "+styleMuted.Render(label))
		}
	} else {
		lines = append(lines, off+" "+name("Timer")+" "+styleMuted.Render("idle"))
	}
	if a.StopwatchRunning {
		lines = append(lines, on+" "+name("Stopwatch")+" "+styleOK.Render(duration.Format(a.StopwatchElapsed)+" elapsed"))
	} else {
		lines = append(lines, off+" "+name("Stopwatch")+" "+styleMuted.Render("idle"))
	}
	if a.PomodoroActive {
		state := fmt.Sprintf("%s %d/%d  %02d:%02d left",
			a.PomodoroPhase, a.PomodoroCycle, a.PomodoroCycles,
			int64(a.PomodoroRemaining/time.Minute),
			int64((a.PomodoroRemaining%time.Minute)/time.Second))
		if a.PomodoroWaiting {
			state = fmt.Sprintf("%s %d/%d  waiting", a.PomodoroPhase, a.PomodoroCycle, a.PomodoroCycles)
		} else if a.PomodoroPaused {
			state = fmt.Sprintf("%s %d/%d  PAUSED  %02d:%02d",
				a.PomodoroPhase, a.PomodoroCycle, a.PomodoroCycles,
				int64(a.PomodoroRemaining/time.Minute),
				int64((a.PomodoroRemaining%time.Minute)/time.Second))
		}
		lines = append(lines, on+" "+name("Pomodoro")+" "+styleOK.Render(state))
	} else {
		lines = append(lines, off+" "+name("Pomodoro")+" "+styleMuted.Render("idle — New Pomodoro or start a preset"))
	}
	return strings.Join(lines, "\n")
}

func renderPresets(items []presetItem, cursor int, focused bool) string {
	if len(items) == 0 {
		return styleMuted.Render("(none — use New Pomodoro / New Timer)")
	}
	var b strings.Builder
	for i, it := range items {
		icon := "◔"
		if it.Kind == kindPomodoro {
			icon = "◷"
		}
		line := fmt.Sprintf("%s %-12s  %s", icon, it.Name, it.Summary)
		switch {
		case focused && i == cursor:
			b.WriteString(styleSel.Render("▸ " + line + " "))
		case i == cursor:
			b.WriteString("▹ " + line)
		default:
			b.WriteString("  " + line)
		}
		b.WriteString("\n")
	}
	return strings.TrimRight(b.String(), "\n")
}
