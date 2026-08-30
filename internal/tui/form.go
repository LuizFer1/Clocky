package tui

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/LuizFer1/Clocky/internal/duration"
	"github.com/LuizFer1/Clocky/internal/presets"
)

type formSavedMsg struct{}
type formCancelMsg struct{}

type formKind int

const (
	formPomodoro formKind = iota
	formTimer
)

type formModel struct {
	root     string
	kind     formKind
	editing  bool
	fields   []string
	labels   []string
	focus    int
	errMsg   string
	title    string
}

func newPomodoroForm(root string, existing *presets.PomodoroPreset) formModel {
	f := formModel{
		root:   root,
		kind:   formPomodoro,
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

func (f formModel) Init() tea.Cmd { return nil }

func (f formModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			return f, func() tea.Msg { return formCancelMsg{} }
		case "enter":
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
	var b strings.Builder
	b.WriteString(styleTitle.Render(f.title))
	b.WriteString("\n\n")
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
		b.WriteString("\n")
	}
	b.WriteString("\n")
	b.WriteString(styleMuted.Render("tab fields  enter save  esc cancel"))
	return b.String()
}
