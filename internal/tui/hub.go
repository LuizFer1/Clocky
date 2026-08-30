package tui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/LuizFer1/Clocky/internal/duration"
	"github.com/LuizFer1/Clocky/internal/pomodoro"
	"github.com/LuizFer1/Clocky/internal/presets"
	"github.com/LuizFer1/Clocky/internal/stopwatch"
	"github.com/LuizFer1/Clocky/internal/timer"
)

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

type hubModel struct {
	deps         Dependencies
	active       activeSnapshot
	items        []presetItem
	cursor       int
	actionCursor int
	focus        hubFocus
	status       string
	errMsg       string
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
		h.active = refreshActive(h.deps.Root, h.deps.Now())
		return h, scheduleTick()
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return h, tea.Quit
		case "tab":
			if h.focus == focusActions {
				h.focus = focusPresets
			} else {
				h.focus = focusActions
			}
			return h, nil
		case "left", "h":
			if h.focus == focusActions && h.actionCursor > 0 {
				h.actionCursor--
			}
			return h, nil
		case "right", "l":
			if h.focus == focusActions && h.actionCursor < len(hubActions)-1 {
				h.actionCursor++
			}
			return h, nil
		case "up", "k":
			if h.focus == focusActions {
				// vertical button layout fallback: move action cursor
				if h.actionCursor > 0 {
					h.actionCursor--
				}
				return h, nil
			}
			if h.cursor > 0 {
				h.cursor--
			}
			return h, nil
		case "down", "j":
			if h.focus == focusActions {
				if h.actionCursor < len(hubActions)-1 {
					h.actionCursor++
				}
				return h, nil
			}
			if h.cursor < len(h.items)-1 {
				h.cursor++
			}
			return h, nil
		case "enter", " ":
			if h.focus == focusActions {
				return h.activateAction(hubActions[h.actionCursor].ID)
			}
			return h.startSelected()
		case "s":
			return h.doStop()
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
	case actionStart:
		return h.startSelected()
	case actionStop:
		return h.doStop()
	case actionStopwatch:
		return h.doStopwatch()
	default:
		return h, nil
	}
}

func (h hubModel) doStop() (tea.Model, tea.Cmd) {
	if err := timer.Stop(h.deps.Root); err != nil {
		h.errMsg = err.Error()
	} else {
		h.status = "Timer stopped"
		h.errMsg = ""
	}
	h.active = refreshActive(h.deps.Root, h.deps.Now())
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
		h.errMsg = "no preset selected — use New Pomodoro / New Timer"
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

	actions := wrapActionBar(h.actionCursor, h.focus == focusActions, panelW, w)
	activePanel := panelBox("Active", renderActive(h.active), panelW)
	presetsTitle := "Presets"
	if h.focus == focusPresets {
		presetsTitle = "Presets  (focused)"
	}
	presetsPanel := panelBox(presetsTitle, renderPresets(h.items, h.cursor, h.focus == focusPresets), panelW)

	var statusLine string
	if h.errMsg != "" {
		statusLine = styleError.Render(h.errMsg)
	} else if h.status != "" {
		statusLine = styleOK.Render(h.status)
	}

	body := joinPanels(actions, activePanel, presetsPanel, statusLine)
	footer := "tab focus  ←→ actions  ↑↓ presets  enter  e edit  d delete  q quit"
	return fillFrame(w, ht, "hub", body, footer)
}

func renderActive(a activeSnapshot) string {
	var lines []string
	if a.TimerActive {
		label := a.TimerLabel
		if label == "" {
			label = "timer"
		}
		lines = append(lines, fmt.Sprintf("[*] Timer      %-12s  %s remaining", label, duration.Format(a.TimerRemaining)))
	} else {
		lines = append(lines, "[ ] Timer      idle")
	}
	if a.StopwatchRunning {
		lines = append(lines, fmt.Sprintf("[*] Stopwatch                %s elapsed", duration.Format(a.StopwatchElapsed)))
	} else {
		lines = append(lines, "[ ] Stopwatch  idle")
	}
	lines = append(lines, "[ ] Pomodoro   idle — New Pomodoro or Start a preset")
	return strings.Join(lines, "\n")
}

func renderPresets(items []presetItem, cursor int, focused bool) string {
	if len(items) == 0 {
		return styleMuted.Render("(none — use New Pomodoro / New Timer)")
	}
	var b strings.Builder
	for i, it := range items {
		prefix := "  "
		line := fmt.Sprintf("%-12s  %s", it.Name, it.Summary)
		if focused && i == cursor {
			prefix = "> "
			b.WriteString(styleSel.Render(prefix + line))
		} else if i == cursor {
			prefix = "· "
			b.WriteString(prefix + line)
		} else {
			b.WriteString(prefix + line)
		}
		b.WriteString("\n")
	}
	return strings.TrimRight(b.String(), "\n")
}
