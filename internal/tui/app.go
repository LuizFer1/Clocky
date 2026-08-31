package tui

import (
	"fmt"
	"os"
	"time"

	"github.com/LuizFer1/Clocky/internal/notify"
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

func (m *appModel) syncHubPomodoro() {
	m.hub.active = mergePomodoroActive(m.hub.active, m.session)
}

func (m appModel) stopSession() appModel {
	m.session.active = false
	m.session.phases = nil
	m.session.status = ""
	m.hub.notice = ""
	m.hub.alerting = false
	notify.StopAlert()
	m.hub.reload()
	m.syncHubPomodoro()
	m.hub.status = "Pomodoro stopped"
	return m
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
		if m.session.live() {
			m.hub.errMsg = "pomodoro already running — Open or Stop it first"
			m.page = pageHub
			return m, scheduleTick()
		}
		m.session = newSessionModel(msg.Cfg, m.width, m.height, nil)
		m.page = pageSession
		return m, m.session.Init()
	case sessionMinimizeMsg:
		m.session.paused = false // background always runs (spec)
		m.page = pageHub
		m.hub.reload()
		m.syncHubPomodoro()
		m.hub.status = "Pomodoro running in background"
		return m, scheduleTick()
	case openPomodoroMsg:
		if !m.session.live() {
			m.hub.errMsg = "no pomodoro running"
			return m, nil
		}
		m.page = pageSession
		return m, scheduleTick()
	case stopPomodoroMsg:
		m = m.stopSession()
		m.page = pageHub
		return m, scheduleTick()
	case sessionDoneMsg:
		m.session.active = false
		m.page = pageHub
		m.hub.reload()
		m.syncHubPomodoro()
		return m, scheduleTick()
	case sessionAbortMsg:
		m = m.stopSession()
		m.page = pageHub
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
	case openLaunchFormMsg:
		switch msg.Kind {
		case formPomodoro:
			m.form = newLaunchPomodoroForm(m.deps.Root)
		case formTimer:
			m.form = newLaunchTimerForm(m.deps.Root)
		}
		m.form.width, m.form.height = m.width, m.height
		m.page = pageForm
		return m, nil
	case formSavedMsg:
		m.page = pageHub
		m.hub.reload()
		m.syncHubPomodoro()
		m.hub.status = "Preset saved"
		m.hub.errMsg = ""
		return m, scheduleTick()
	case formLaunchPomodoroMsg:
		if m.session.live() {
			m.hub.errMsg = "pomodoro already running — Open or Stop it first"
			m.page = pageHub
			m.hub.reload()
			m.syncHubPomodoro()
			return m, scheduleTick()
		}
		m.page = pageHub
		m.hub.reload()
		m.syncHubPomodoro()
		if msg.SavedPreset {
			m.hub.status = "Preset saved · starting pomodoro"
		} else {
			m.hub.status = "Starting pomodoro"
		}
		m.hub.errMsg = ""
		m.session = newSessionModel(msg.Cfg, m.width, m.height, nil)
		m.page = pageSession
		return m, m.session.Init()
	case formLaunchTimerMsg:
		m.page = pageHub
		m.hub.reload()
		m.syncHubPomodoro()
		mod, _ := m.hub.startTimer(msg.Duration, msg.Label)
		m.hub = mod.(hubModel)
		m.syncHubPomodoro()
		if msg.SavedPreset {
			m.hub.status = fmt.Sprintf("Preset saved · timer started: %s", msg.Label)
		}
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
		m.syncHubPomodoro()
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
		m.syncHubPomodoro()
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
