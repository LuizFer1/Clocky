package tui

import (
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type page int

const (
	pageHub page = iota
	pageSession
	pageForm
	pageConfirm
	pagePicker
)

// Dependencies injects filesystem root and clock for tests.
type Dependencies struct {
	Root string
	Now  func() time.Time
	Exe  func() (string, error)
}

func (d Dependencies) withDefaults() Dependencies {
	if d.Now == nil {
		d.Now = time.Now
	}
	if d.Exe == nil {
		d.Exe = os.Executable
	}
	return d
}

type appModel struct {
	deps     Dependencies
	page     page
	hub      hubModel
	session  sessionModel
	form     formModel
	confirm  confirmModel
	picker   pickerModel
	width    int
	height   int
	pendingKind string
	pendingName string
}

func initialModel(deps Dependencies) appModel {
	deps = deps.withDefaults()
	return appModel{
		deps: deps,
		page: pageHub,
		hub:  newHubModel(deps),
	}
}

func (m appModel) Init() tea.Cmd {
	return m.hub.Init()
}

func (m appModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.hub.width = msg.Width
		m.hub.height = msg.Height
		m.session.width = msg.Width
		m.session.height = msg.Height
		m.form.width = msg.Width
		m.form.height = msg.Height
		m.confirm.width = msg.Width
		m.confirm.height = msg.Height
		m.picker.width = msg.Width
		m.picker.height = msg.Height
		return m, nil
	case startSessionMsg:
		m.session = newSessionModel(msg.Cfg, m.width, m.height, nil)
		m.page = pageSession
		return m, m.session.Init()
	case sessionDoneMsg, sessionAbortMsg:
		m.page = pageHub
		m.hub.reload()
		m.hub.status = "Back at hub"
		return m, scheduleTick()
	case openPickerMsg:
		m.picker = newPickerModel()
		m.picker.width, m.picker.height = m.width, m.height
		m.page = pagePicker
		return m, nil
	case pickerCancelMsg:
		m.page = pageHub
		return m, scheduleTick()
	case pickerChoiceMsg:
		switch msg.Kind {
		case formPomodoro:
			m.form = newPomodoroForm(m.deps.Root, nil)
		case formTimer:
			m.form = newTimerForm(m.deps.Root, nil)
		}
		m.form.width, m.form.height = m.width, m.height
		m.page = pageForm
		return m, nil
	case openFormMsg:
		switch msg.Kind {
		case formPomodoro:
			m.form = newPomodoroForm(m.deps.Root, msg.Pomo)
		case formTimer:
			m.form = newTimerForm(m.deps.Root, msg.Tim)
		}
		m.form.width, m.form.height = m.width, m.height
		m.page = pageForm
		return m, nil
	case formSavedMsg:
		m.page = pageHub
		m.hub.reload()
		m.hub.status = "Preset saved"
		m.hub.errMsg = ""
		return m, scheduleTick()
	case formCancelMsg:
		m.page = pageHub
		return m, scheduleTick()
	case openConfirmMsg:
		m.pendingKind = msg.Kind
		m.pendingName = msg.Name
		m.confirm = newConfirmModel(msg.Kind, msg.Name)
		m.confirm.width, m.confirm.height = m.width, m.height
		m.page = pageConfirm
		return m, nil
	case confirmResultMsg:
		if msg.OK {
			if err := deletePreset(m.deps.Root, m.pendingKind, m.pendingName); err != nil {
				m.hub.errMsg = err.Error()
			} else {
				m.hub.status = "Preset deleted"
				m.hub.errMsg = ""
			}
		}
		m.page = pageHub
		m.hub.reload()
		return m, scheduleTick()
	}

	switch m.page {
	case pageHub:
		mod, cmd := m.hub.Update(msg)
		m.hub = mod.(hubModel)
		return m, cmd
	case pageSession:
		mod, cmd := m.session.Update(msg)
		m.session = mod.(sessionModel)
		return m, cmd
	case pageForm:
		mod, cmd := m.form.Update(msg)
		m.form = mod.(formModel)
		return m, cmd
	case pageConfirm:
		mod, cmd := m.confirm.Update(msg)
		m.confirm = mod.(confirmModel)
		return m, cmd
	case pagePicker:
		mod, cmd := m.picker.Update(msg)
		m.picker = mod.(pickerModel)
		return m, cmd
	}
	return m, nil
}

func (m appModel) View() string {
	switch m.page {
	case pageSession:
		return m.session.View()
	case pageForm:
		return m.form.View()
	case pageConfirm:
		return m.confirm.View()
	case pagePicker:
		return m.picker.View()
	default:
		return m.hub.View()
	}
}
