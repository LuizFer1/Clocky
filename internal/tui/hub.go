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

type openConfirmMsg struct {
	Kind string
	Name string
}

type hubModel struct {
	deps   Dependencies
	active activeSnapshot
	items  []presetItem
	cursor int
	status string
	errMsg string
	width  int
}

func newHubModel(deps Dependencies) hubModel {
	h := hubModel{deps: deps, width: 80}
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
		return h, nil
	case tickMsg:
		h.active = refreshActive(h.deps.Root, h.deps.Now())
		return h, scheduleTick()
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return h, tea.Quit
		case "up", "k":
			if h.cursor > 0 {
				h.cursor--
			}
			return h, nil
		case "down", "j":
			if h.cursor < len(h.items)-1 {
				h.cursor++
			}
			return h, nil
		case "s":
			if err := timer.Stop(h.deps.Root); err != nil {
				h.errMsg = err.Error()
			} else {
				h.status = "Timer stopped"
				h.errMsg = ""
			}
			h.active = refreshActive(h.deps.Root, h.deps.Now())
			return h, nil
		case "t":
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
		case "p":
			return h, func() tea.Msg {
				return startSessionMsg{Cfg: pomodoro.Config{
					Focus: 25 * time.Minute, Break: 5 * time.Minute,
					Long: 15 * time.Minute, Cycles: 4, Auto: false,
				}}
			}
		case "n":
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
		case "enter":
			return h.startSelected()
		}
	}
	return h, nil
}

func (h hubModel) startSelected() (tea.Model, tea.Cmd) {
	if len(h.items) == 0 {
		h.errMsg = "no presets — press n to create"
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
		exe, err := h.deps.Exe()
		if err != nil {
			h.errMsg = err.Error()
			return h, nil
		}
		d := time.Duration(it.Tim.Seconds) * time.Second
		if err := timer.Start(h.deps.Root, d, it.Name, exe, []string{"timer", "--worker"}); err != nil {
			h.errMsg = err.Error()
			return h, nil
		}
		h.status = fmt.Sprintf("Timer started: %s (%s)", it.Name, duration.Format(d))
		h.errMsg = ""
		h.active = refreshActive(h.deps.Root, h.deps.Now())
		return h, nil
	}
	return h, nil
}

func (h hubModel) View() string {
	w := h.width
	if w <= 0 {
		w = 80
	}
	inner := w - 4
	if inner < 20 {
		inner = 20
	}

	activePanel := stylePanel.Width(inner).Render(styleTitle.Render("Active") + "\n" + renderActive(h.active))
	presetsPanel := stylePanel.Width(inner).Render(styleTitle.Render("Presets") + "\n" + renderPresets(h.items, h.cursor))

	var b strings.Builder
	b.WriteString(styleTitle.Render("Clocky"))
	b.WriteString("\n")
	b.WriteString(activePanel)
	b.WriteString("\n")
	b.WriteString(presetsPanel)
	b.WriteString("\n")
	if h.errMsg != "" {
		b.WriteString(styleError.Render(h.errMsg))
		b.WriteString("\n")
	} else if h.status != "" {
		b.WriteString(styleOK.Render(h.status))
		b.WriteString("\n")
	}
	b.WriteString(styleMuted.Render("↑↓ select  ↵ start  n new  e edit  d delete  s stop timer  t stopwatch  p pomodoro  q quit"))
	return b.String()
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
	lines = append(lines, "[ ] Pomodoro   idle (p or Enter on preset)")
	return strings.Join(lines, "\n")
}

func renderPresets(items []presetItem, cursor int) string {
	if len(items) == 0 {
		return styleMuted.Render("(none — press n to create)")
	}
	var b strings.Builder
	for i, it := range items {
		prefix := "  "
		line := fmt.Sprintf("%-12s  %s", it.Name, it.Summary)
		if i == cursor {
			prefix = "> "
			b.WriteString(styleSel.Render(prefix + line))
		} else {
			b.WriteString(prefix + line)
		}
		b.WriteString("\n")
	}
	return strings.TrimRight(b.String(), "\n")
}
