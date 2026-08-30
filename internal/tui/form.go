package tui

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/LuizFer1/Clocky/internal/duration"
	"github.com/LuizFer1/Clocky/internal/pomodoro"
	"github.com/LuizFer1/Clocky/internal/presets"
)

type formSavedMsg struct{}
type formCancelMsg struct{}

// formLaunchPomodoroMsg starts a session (and may have saved a preset).
type formLaunchPomodoroMsg struct {
	Cfg        pomodoro.Config
	SavedName  string
	SavedPreset bool
}

// formLaunchTimerMsg starts a background timer (and may have saved a preset).
type formLaunchTimerMsg struct {
	Duration    time.Duration
	Label       string
	SavedName   string
	SavedPreset bool
}

type formKind int

const (
	formPomodoro formKind = iota
	formTimer
)

type formMode int

const (
	formModePreset formMode = iota // create/edit preset only
	formModeLaunch                // configure + start (+ optional save)
)

type formModel struct {
	root    string
	kind    formKind
	mode    formMode
	editing bool
	fields  []string
	labels  []string
	focus   int
	errMsg  string
	title   string
	width   int
	height  int
}

func newPomodoroForm(root string, existing *presets.PomodoroPreset) formModel {
	f := formModel{
		root:   root,
		kind:   formPomodoro,
		mode:   formModePreset,
		title:  "New pomodoro preset",
		labels: []string{"Name", "Focus (min)", "Break (min)", "Long (min)", "Cycles", "Auto (y/n)"},
		fields: []string{"", "25", "5", "15", "4", "n"},
	}
	if existing != nil {
		f.editing = true
		f.title = "Edit pomodoro preset"
		auto := "n"
		if existing.Auto {
			auto = "y"
		}
		f.fields = []string{
			existing.Name,
			strconv.Itoa(existing.Focus),
			strconv.Itoa(existing.Break),
			strconv.Itoa(existing.Long),
			strconv.Itoa(existing.Cycles),
			auto,
		}
	}
	return f
}

func newTimerForm(root string, existing *presets.TimerPreset) formModel {
	f := formModel{
		root:   root,
		kind:   formTimer,
		mode:   formModePreset,
		title:  "New timer preset",
		labels: []string{"Name", "Duration (H:M:S)"},
		fields: []string{"", ":25:"},
	}
	if existing != nil {
		f.editing = true
		f.title = "Edit timer preset"
		f.fields = []string{
			existing.Name,
			duration.Format(time.Duration(existing.Seconds) * time.Second),
		}
	}
	return f
}

// newLaunchPomodoroForm configures a pomodoro then starts it.
// Optional name saves a preset when non-empty.
func newLaunchPomodoroForm(root string) formModel {
	return formModel{
		root:  root,
		kind:  formPomodoro,
		mode:  formModeLaunch,
		title: "New Pomodoro",
		labels: []string{
			"Focus (min)",
			"Break (min)",
			"Long break (min)",
			"Cycles",
			"Auto (y/n)",
			"Save as preset (optional name)",
		},
		fields: []string{"25", "5", "15", "4", "n", ""},
	}
}

// newLaunchTimerForm configures a timer then starts it.
func newLaunchTimerForm(root string) formModel {
	return formModel{
		root:  root,
		kind:  formTimer,
		mode:  formModeLaunch,
		title: "New Timer",
		labels: []string{
			"Duration (H:M:S)",
			"Label (optional)",
			"Save as preset (optional name)",
		},
		fields: []string{":25:", "", ""},
	}
}

func (f formModel) Init() tea.Cmd { return nil }

func (f formModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			return f, func() tea.Msg { return formCancelMsg{} }
		case "enter":
			if f.mode == formModeLaunch {
				cmd, err := f.submitLaunch()
				if err != nil {
					f.errMsg = err.Error()
					return f, nil
				}
				return f, cmd
			}
			if err := f.save(); err != nil {
				f.errMsg = err.Error()
				return f, nil
			}
			return f, func() tea.Msg { return formSavedMsg{} }
		case "tab", "down":
			f.focus = (f.focus + 1) % len(f.fields)
			return f, nil
		case "shift+tab", "up":
			f.focus = (f.focus - 1 + len(f.fields)) % len(f.fields)
			return f, nil
		case "backspace":
			if f.focus < len(f.fields) && len(f.fields[f.focus]) > 0 {
				f.fields[f.focus] = f.fields[f.focus][:len(f.fields[f.focus])-1]
			}
			return f, nil
		case "ctrl+c", "q":
			return f, tea.Quit
		default:
			if msg.Type == tea.KeyRunes {
				f.fields[f.focus] += string(msg.Runes)
			}
			return f, nil
		}
	}
	return f, nil
}

func (f formModel) submitLaunch() (tea.Cmd, error) {
	switch f.kind {
	case formPomodoro:
		focus, err := strconv.Atoi(strings.TrimSpace(f.fields[0]))
		if err != nil || focus <= 0 {
			return nil, fmt.Errorf("invalid focus minutes")
		}
		brk, err := strconv.Atoi(strings.TrimSpace(f.fields[1]))
		if err != nil || brk <= 0 {
			return nil, fmt.Errorf("invalid break minutes")
		}
		long, err := strconv.Atoi(strings.TrimSpace(f.fields[2]))
		if err != nil || long <= 0 {
			return nil, fmt.Errorf("invalid long minutes")
		}
		cycles, err := strconv.Atoi(strings.TrimSpace(f.fields[3]))
		if err != nil || cycles < 1 {
			return nil, fmt.Errorf("cycles must be >= 1")
		}
		autoRaw := strings.ToLower(strings.TrimSpace(f.fields[4]))
		auto := autoRaw == "y" || autoRaw == "yes" || autoRaw == "true" || autoRaw == "1"
		saveName := strings.TrimSpace(f.fields[5])
		saved := false
		if saveName != "" {
			if err := savePomodoroPreset(f.root, presets.PomodoroPreset{
				Name: saveName, Focus: focus, Break: brk, Long: long, Cycles: cycles, Auto: auto,
			}); err != nil {
				return nil, err
			}
			saved = true
		}
		msg := formLaunchPomodoroMsg{
			Cfg: pomodoro.Config{
				Focus:  time.Duration(focus) * time.Minute,
				Break:  time.Duration(brk) * time.Minute,
				Long:   time.Duration(long) * time.Minute,
				Cycles: cycles,
				Auto:   auto,
			},
			SavedName:   saveName,
			SavedPreset: saved,
		}
		return func() tea.Msg { return msg }, nil
	case formTimer:
		d, err := duration.Parse(strings.TrimSpace(f.fields[0]))
		if err != nil {
			return nil, err
		}
		label := strings.TrimSpace(f.fields[1])
		saveName := strings.TrimSpace(f.fields[2])
		saved := false
		if saveName != "" {
			if err := saveTimerPreset(f.root, presets.TimerPreset{
				Name: saveName, Seconds: int64(d / time.Second),
			}); err != nil {
				return nil, err
			}
			saved = true
			if label == "" {
				label = saveName
			}
		}
		if label == "" {
			label = duration.Format(d)
		}
		msg := formLaunchTimerMsg{
			Duration: d, Label: label, SavedName: saveName, SavedPreset: saved,
		}
		return func() tea.Msg { return msg }, nil
	default:
		return nil, fmt.Errorf("unknown form")
	}
}

func (f formModel) save() error {
	switch f.kind {
	case formPomodoro:
		focus, err := strconv.Atoi(strings.TrimSpace(f.fields[1]))
		if err != nil {
			return fmt.Errorf("invalid focus minutes")
		}
		brk, err := strconv.Atoi(strings.TrimSpace(f.fields[2]))
		if err != nil {
			return fmt.Errorf("invalid break minutes")
		}
		long, err := strconv.Atoi(strings.TrimSpace(f.fields[3]))
		if err != nil {
			return fmt.Errorf("invalid long minutes")
		}
		cycles, err := strconv.Atoi(strings.TrimSpace(f.fields[4]))
		if err != nil {
			return fmt.Errorf("invalid cycles")
		}
		autoRaw := strings.ToLower(strings.TrimSpace(f.fields[5]))
		auto := autoRaw == "y" || autoRaw == "yes" || autoRaw == "true" || autoRaw == "1"
		return savePomodoroPreset(f.root, presets.PomodoroPreset{
			Name:   f.fields[0],
			Focus:  focus,
			Break:  brk,
			Long:   long,
			Cycles: cycles,
			Auto:   auto,
		})
	case formTimer:
		d, err := duration.Parse(strings.TrimSpace(f.fields[1]))
		if err != nil {
			return err
		}
		return saveTimerPreset(f.root, presets.TimerPreset{
			Name:    f.fields[0],
			Seconds: int64(d / time.Second),
		})
	default:
		return fmt.Errorf("unknown form")
	}
}

func (f formModel) View() string {
	w, ht := f.width, f.height
	if w <= 0 {
		w = 80
	}
	if ht <= 0 {
		ht = 24
	}
	panelW := min(56, max(28, w-10))
	var b strings.Builder
	for i, label := range f.labels {
		cursor := "  "
		val := f.fields[i]
		if i == f.focus {
			cursor = "> "
			val = val + "█"
		}
		line := fmt.Sprintf("%s%s: %s", cursor, label, val)
		if i == f.focus {
			b.WriteString(styleSel.Render(line))
		} else {
			b.WriteString(line)
		}
		b.WriteString("\n")
	}
	if f.errMsg != "" {
		b.WriteString("\n")
		b.WriteString(styleError.Render(f.errMsg))
	}
	body := panelBox(f.title, strings.TrimRight(b.String(), "\n"), panelW)
	sub := "Clocky  ·  preset"
	foot := "tab fields  enter save  esc cancel"
	if f.mode == formModeLaunch {
		sub = "Clocky  ·  start"
		foot = "tab fields  enter start  esc cancel"
	}
	return fillFrame(w, ht, sub, body, foot)
}
