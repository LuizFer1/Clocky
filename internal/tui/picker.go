package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type pickerChoiceMsg struct {
	Kind formKind
}

type pickerCancelMsg struct{}

type pickerModel struct {
	cursor int
}

func newPickerModel() pickerModel { return pickerModel{} }

func (p pickerModel) Init() tea.Cmd { return nil }

func (p pickerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if p.cursor > 0 {
				p.cursor--
			}
			return p, nil
		case "down", "j":
			if p.cursor < 1 {
				p.cursor++
			}
			return p, nil
		case "enter":
			kind := formPomodoro
			if p.cursor == 1 {
				kind = formTimer
			}
			return p, func() tea.Msg { return pickerChoiceMsg{Kind: kind} }
		case "esc":
			return p, func() tea.Msg { return pickerCancelMsg{} }
		case "q", "ctrl+c":
			return p, tea.Quit
		}
	}
	return p, nil
}

func (p pickerModel) View() string {
	opts := []string{"Pomodoro preset", "Timer preset"}
	var b strings.Builder
	b.WriteString(styleTitle.Render("New preset"))
	b.WriteString("\n\n")
	for i, o := range opts {
		cur := "  "
		line := o
		if i == p.cursor {
			cur = "> "
			line = styleSel.Render(o)
		}
		b.WriteString(cur)
		b.WriteString(line)
		b.WriteString("\n")
	}
	b.WriteString("\n")
	b.WriteString(styleMuted.Render("enter choose  esc cancel"))
	return b.String()
}
