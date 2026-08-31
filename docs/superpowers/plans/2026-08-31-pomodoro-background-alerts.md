# Pomodoro Background + Alerts Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Keep the TUI pomodoro running when the user returns to the hub, fix phase-end visual corruption, and alert with the same notice + repeating beep pattern as finished timers.

**Architecture:** Own one live `sessionModel` on `appModel` for the TUI process lifetime. Esc minimizes to the hub (does not cancel). Hub ticks advance the session when the session page is not focused. Phase end never writes a stdout banner; it sets hub notice + `notify.Alert()` + desktop toast.

**Tech Stack:** Go, Bubble Tea, Lip Gloss; existing `internal/tui`, `internal/pomodoro`, `internal/notify`.

**Spec:** `docs/superpowers/specs/2026-08-31-pomodoro-background-alerts-design.md`

---

## File map

| Path | Responsibility |
|------|----------------|
| `internal/tui/session.go` | Phase tick / finish without `notify.All`; minimize msg; snapshot helper |
| `internal/tui/active.go` | Extend `activeSnapshot` with pomodoro fields |
| `internal/tui/buttons.go` | Open Pomodoro / Stop Pomodoro actions when session live |
| `internal/tui/hub.go` | Render pomodoro Active line; handle Open/Stop; Enter advances when waiting |
| `internal/tui/app.go` | Lifecycle: minimize / open / stop / refuse double start; hub-side tick; alert pipeline |
| `internal/tui/notice.go` | Optional helper for pomodoro phase notice text (keep with timer notices) |
| `internal/tui/*_test.go` | Lifecycle, background tick, no-banner, buttons, Active render |

---

### Task 1: Phase-end event without stdout Banner

**Files:**
- Modify: `internal/tui/session.go`
- Test: `internal/tui/session_test.go`

- [ ] **Step 1: Write the failing test**

Append to `internal/tui/session_test.go`:

```go
func TestFinishPhaseDoesNotCallBanner(t *testing.T) {
	rec := &notify.RecordingNotifier{}
	cfg := pomodoro.Config{
		Focus:  time.Second,
		Break:  time.Second,
		Long:   time.Second,
		Cycles: 1,
		Auto:   true,
	}
	s := newSessionModel(cfg, 40, 24, rec)
	s.remaining = 0
	m, cmd := s.finishPhase()
	s = m.(sessionModel)
	_ = s
	if cmd == nil {
		t.Fatal("expected command from finishPhase")
	}
	msg := cmd()
	end, ok := msg.(sessionPhaseEndMsg)
	if !ok {
		t.Fatalf("got %#v want sessionPhaseEndMsg", msg)
	}
	if end.Title != "Focus complete" {
		t.Fatalf("title=%q", end.Title)
	}
	for _, e := range rec.Events {
		if strings.HasPrefix(e, "banner:") {
			t.Fatalf("banner must not be called from TUI finishPhase: %v", rec.Events)
		}
	}
}
```

Add imports: `"strings"`, `"github.com/LuizFer1/Clocky/internal/notify"`.

- [ ] **Step 2: Run test to verify it fails**

Run:

```powershell
go test ./internal/tui/ -run TestFinishPhaseDoesNotCallBanner -count=1
```

Expected: FAIL (missing `sessionPhaseEndMsg` and/or still calling `notify.All`).

- [ ] **Step 3: Minimal implementation**

In `internal/tui/session.go`:

1. Replace `sessionAbortMsg` usage for Esc with minimize (abort type can remain unused or be removed later). Add:

```go
type sessionMinimizeMsg struct{}

type sessionPhaseEndMsg struct {
	Title string
	Body  string
	Done  bool // true when the whole pomodoro finished
}
```

2. Change `finishPhase` to **not** call `notify.All`. Emit `sessionPhaseEndMsg` instead:

```go
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
		return s, tea.Batch(phaseCmd, scheduleTick())
	}
	// Auto: advance, then notify about the phase that just ended.
	m, advCmd := s.advanceAfterWait()
	return m, tea.Batch(phaseCmd, advCmd)
}
```

3. Esc / `b` minimize:

```go
case "esc", "b":
	return s, func() tea.Msg { return sessionMinimizeMsg{} }
```

Keep `notifier` field for now (desktop may be triggered from app); session itself must not Banner/Beep via `notify.All`.

- [ ] **Step 4: Run test to verify it passes**

```powershell
go test ./internal/tui/ -run TestFinishPhaseDoesNotCallBanner -count=1
```

Expected: PASS.

- [ ] **Step 5: Commit**

```powershell
git add internal/tui/session.go internal/tui/session_test.go
git commit -m "fix(tui): emit phase-end msg without stdout banner"
```

---

### Task 2: Pomodoro fields on Active + Open/Stop buttons

**Files:**
- Modify: `internal/tui/active.go`
- Modify: `internal/tui/hub.go` (`renderActive`)
- Modify: `internal/tui/buttons.go`
- Modify: `internal/tui/hub.go` (`activateAction`, helpers)
- Test: `internal/tui/buttons_test.go`, `internal/tui/active_test.go` (create if missing patterns)

- [ ] **Step 1: Write failing tests**

In `internal/tui/buttons_test.go`:

```go
func TestHubActionButtonsIncludePomodoroControls(t *testing.T) {
	a := activeSnapshot{PomodoroActive: true}
	btns := hubActionButtons(a)
	labels := make([]string, len(btns))
	for i, b := range btns {
		labels[i] = b.Label
	}
	joined := strings.Join(labels, ",")
	if !strings.Contains(joined, "Open Pomodoro") || !strings.Contains(joined, "Stop Pomodoro") {
		t.Fatalf("buttons=%v", labels)
	}
}
```

In `internal/tui/hub_test.go` (or new `active_test.go`):

```go
func TestRenderActiveShowsPomodoro(t *testing.T) {
	a := activeSnapshot{
		PomodoroActive:    true,
		PomodoroPhase:     "FOCUS",
		PomodoroCycle:     1,
		PomodoroCycles:    4,
		PomodoroRemaining: 12*time.Minute + 34*time.Second,
	}
	got := renderActive(a)
	if !strings.Contains(got, "FOCUS") || !strings.Contains(got, "1/4") {
		t.Fatalf("active=%q", got)
	}
	if !strings.Contains(got, "12:34") && !strings.Contains(got, "12m") {
		// duration.Format or MM:SS — accept either if you standardize on MM:SS in render
		t.Fatalf("missing remaining in %q", got)
	}
}
```

Use a consistent format in implementation: `FOCUS 1/4  12:34 left` (MM:SS), matching the session compact line.

- [ ] **Step 2: Run tests — expect FAIL**

```powershell
go test ./internal/tui/ -run "TestHubActionButtonsIncludePomodoroControls|TestRenderActiveShowsPomodoro" -count=1
```

- [ ] **Step 3: Extend snapshot + render + buttons**

`internal/tui/active.go`:

```go
type activeSnapshot struct {
	TimerActive      bool
	TimerRemaining   time.Duration
	TimerLabel       string
	StopwatchRunning bool
	StopwatchElapsed time.Duration

	PomodoroActive    bool
	PomodoroPhase     string
	PomodoroCycle     int
	PomodoroCycles    int
	PomodoroRemaining time.Duration
	PomodoroWaiting   bool
	PomodoroPaused    bool
}
```

`refreshActive` stays disk-only for timer/stopwatch. Pomodoro fields are merged by the app (Task 3) via:

```go
func mergePomodoroActive(a activeSnapshot, s sessionModel) activeSnapshot {
	if !s.live() {
		return a
	}
	a.PomodoroActive = true
	if len(s.phases) == 0 {
		return a
	}
	ph := s.phases[s.index]
	a.PomodoroPhase = ph.Name
	a.PomodoroCycle = ph.Cycle
	a.PomodoroCycles = s.cfg.Cycles
	a.PomodoroRemaining = s.remaining
	a.PomodoroWaiting = s.waitingEnter
	a.PomodoroPaused = s.paused
	return a
}
```

Add on `sessionModel`:

```go
func (s sessionModel) live() bool {
	return len(s.phases) > 0 && s.index < len(s.phases) && s.status != "stopped"
}
```

Use an explicit `active bool` field set true in `newSessionModel` and false on stop/done — clearer than status string:

```go
// in sessionModel:
active bool

func newSessionModel(...) sessionModel {
	// ...
	s.active = len(phases) > 0
	return s
}

func (s sessionModel) live() bool { return s.active }
```

`renderActive` — replace the idle-only pomodoro line:

```go
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
```

`buttons.go`:

```go
const (
	actionNewPomodoro actionID = iota
	actionNewTimer
	actionStopwatchToggle
	actionStopTimer
	actionOpenPomodoro
	actionStopPomodoro
)

func hubActionButtons(a activeSnapshot) []actionButton {
	swLabel := "Start Stopwatch"
	if a.StopwatchRunning {
		swLabel = "Stop Stopwatch"
	}
	var btns []actionButton
	if a.PomodoroActive {
		btns = append(btns,
			actionButton{actionOpenPomodoro, "Open Pomodoro"},
			actionButton{actionStopPomodoro, "Stop Pomodoro"},
		)
	} else {
		btns = append(btns, actionButton{actionNewPomodoro, "New Pomodoro"})
	}
	btns = append(btns,
		actionButton{actionNewTimer, "New Timer"},
		actionButton{actionStopwatchToggle, swLabel},
	)
	if a.TimerActive {
		btns = append(btns, actionButton{actionStopTimer, "Stop Timer"})
	}
	return btns
}
```

In `hub.go` `activateAction`, handle new IDs by returning msgs:

```go
type openPomodoroMsg struct{}
type stopPomodoroMsg struct{}

case actionOpenPomodoro:
	return h, func() tea.Msg { return openPomodoroMsg{} }
case actionStopPomodoro:
	return h, func() tea.Msg { return stopPomodoroMsg{} }
```

- [ ] **Step 4: Run tests — expect PASS**

```powershell
go test ./internal/tui/ -run "TestHubActionButtonsIncludePomodoroControls|TestRenderActiveShowsPomodoro|TestHubActionBarInView|TestActivateNewPomodoroOpensLaunchForm" -count=1
```

Fix `TestActivateNewPomodoroOpensLaunchForm` if button index shifted when idle (still index 0 = New Pomodoro).

- [ ] **Step 5: Commit**

```powershell
git add internal/tui/active.go internal/tui/buttons.go internal/tui/buttons_test.go internal/tui/hub.go internal/tui/hub_test.go internal/tui/session.go
git commit -m "feat(tui): show live pomodoro on hub Active with Open/Stop"
```

---

### Task 3: App lifecycle — minimize, open, stop, refuse double start

**Files:**
- Modify: `internal/tui/app.go`
- Test: `internal/tui/app_test.go`

- [ ] **Step 1: Write failing tests**

```go
func TestMinimizeKeepsSession(t *testing.T) {
	m := initialModel(Dependencies{Root: t.TempDir(), Now: time.Now})
	m.session = newSessionModel(pomodoro.Config{
		Focus: 5 * time.Minute, Break: time.Minute, Long: time.Minute, Cycles: 2, Auto: true,
	}, 80, 24, nil)
	m.page = pageSession
	before := m.session.remaining
	mod, _ := m.Update(sessionMinimizeMsg{})
	m = mod.(appModel)
	if m.page != pageHub {
		t.Fatalf("page=%v", m.page)
	}
	if !m.session.live() || m.session.remaining != before {
		t.Fatalf("session cleared or changed: live=%v rem=%v", m.session.live(), m.session.remaining)
	}
}

func TestStopPomodoroClearsSession(t *testing.T) {
	m := initialModel(Dependencies{Root: t.TempDir(), Now: time.Now})
	m.session = newSessionModel(pomodoro.Config{
		Focus: time.Minute, Break: time.Minute, Long: time.Minute, Cycles: 1, Auto: true,
	}, 80, 24, nil)
	m.page = pageHub
	m.syncHubPomodoro()
	mod, _ := m.Update(stopPomodoroMsg{})
	m = mod.(appModel)
	if m.session.live() {
		t.Fatal("expected stopped")
	}
	if m.hub.active.PomodoroActive {
		t.Fatal("active still shows pomodoro")
	}
}

func TestRefuseSecondPomodoroStart(t *testing.T) {
	m := initialModel(Dependencies{Root: t.TempDir(), Now: time.Now})
	m.session = newSessionModel(pomodoro.Config{
		Focus: time.Minute, Break: time.Minute, Long: time.Minute, Cycles: 1, Auto: true,
	}, 80, 24, nil)
	m.page = pageHub
	mod, _ := m.Update(startSessionMsg{Cfg: pomodoro.Config{
		Focus: time.Minute, Break: time.Minute, Long: time.Minute, Cycles: 1, Auto: true,
	}})
	m = mod.(appModel)
	if m.page == pageSession {
		t.Fatal("should refuse starting a second session")
	}
	if m.hub.errMsg == "" {
		t.Fatal("expected error message")
	}
}
```

Add `syncHubPomodoro` in implementation (Step 3). Import `pomodoro` in `app_test.go` if missing.

- [ ] **Step 2: Run — expect FAIL**

```powershell
go test ./internal/tui/ -run "TestMinimizeKeepsSession|TestStopPomodoroClearsSession|TestRefuseSecondPomodoroStart" -count=1
```

- [ ] **Step 3: Wire app.go**

```go
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
```

Update `Update` handlers:

```go
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
	// Phase-end alert already applied via sessionPhaseEndMsg.
	m.session.active = false
	m.page = pageHub
	m.hub.reload()
	m.syncHubPomodoro()
	return m, scheduleTick()

case sessionAbortMsg:
	// Legacy path: treat as stop if any test still sends it.
	m = m.stopSession()
	m.page = pageHub
	return m, scheduleTick()
```

Apply the same `live()` guard to `formLaunchPomodoroMsg` before creating a new session.

After any hub reload while session live, call `syncHubPomodoro()` so disk refresh does not wipe pomodoro fields — implement `merge` after `reload`:

```go
func (h *hubModel) reload() {
	// existing body...
}

// in app after hub.reload():
m.hub.reload()
m.syncHubPomodoro()
```

- [ ] **Step 4: Run — expect PASS**

```powershell
go test ./internal/tui/ -run "TestMinimizeKeepsSession|TestStopPomodoroClearsSession|TestRefuseSecondPomodoroStart" -count=1
```

- [ ] **Step 5: Commit**

```powershell
git add internal/tui/app.go internal/tui/app_test.go internal/tui/session.go
git commit -m "feat(tui): minimize/open/stop pomodoro without destroying session"
```

---

### Task 4: Background hub tick + alert pipeline (Desktop + notice + Alert)

**Files:**
- Modify: `internal/tui/app.go`
- Modify: `internal/tui/session.go` (ensure tick works when driven from app)
- Modify: `internal/tui/hub.go` only if needed for Enter-to-advance
- Test: `internal/tui/app_test.go`

- [ ] **Step 1: Write failing tests**

```go
type recordDesktop struct {
	notify.RecordingNotifier
}

func TestHubTickAdvancesBackgroundPomodoro(t *testing.T) {
	m := initialModel(Dependencies{Root: t.TempDir(), Now: time.Now})
	m.session = newSessionModel(pomodoro.Config{
		Focus: 3 * time.Second, Break: time.Second, Long: time.Second, Cycles: 1, Auto: true,
	}, 80, 24, nil)
	m.page = pageHub
	m.syncHubPomodoro()
	before := m.session.remaining
	mod, _ := m.Update(tickMsg{})
	m = mod.(appModel)
	if m.session.remaining != before-time.Second {
		t.Fatalf("remaining=%v want %v", m.session.remaining, before-time.Second)
	}
	if !m.hub.active.PomodoroActive {
		t.Fatal("hub active lost pomodoro")
	}
}

func TestPhaseEndOnHubSetsNoticeAndDesktop(t *testing.T) {
	rec := &notify.RecordingNotifier{}
	m := initialModel(Dependencies{Root: t.TempDir(), Now: time.Now})
	m.session = newSessionModel(pomodoro.Config{
		Focus: time.Second, Break: time.Second, Long: time.Second, Cycles: 2, Auto: false,
	}, 80, 24, rec)
	m.page = pageHub
	m.session.remaining = time.Second
	// Two ticks: first -> 0 triggers finish; or set remaining=0 and tick
	m.session.remaining = 0
	mod, cmd := m.Update(tickMsg{})
	m = mod.(appModel)
	// Drain phase-end cmd if returned as Batch — call cmd and feed msgs back
	if cmd != nil {
		msg := cmd()
		// tea.Batch returns a batchMsg — handle by applying sessionPhaseEndMsg directly in test helper
		_ = msg
	}
	// Prefer testing the handler directly:
	mod, cmd = m.Update(sessionPhaseEndMsg{Title: "Focus complete", Body: "Focus session 1 finished", Done: false})
	m = mod.(appModel)
	if m.hub.notice == "" || !m.hub.alerting {
		t.Fatalf("notice=%q alerting=%v", m.hub.notice, m.hub.alerting)
	}
	if !strings.Contains(m.hub.notice, "Focus complete") {
		t.Fatalf("notice=%q", m.hub.notice)
	}
	hasDesktop := false
	for _, e := range rec.Events {
		if strings.HasPrefix(e, "desktop:") {
			hasDesktop = true
		}
		if strings.HasPrefix(e, "banner:") {
			t.Fatalf("banner leaked: %v", rec.Events)
		}
	}
	if !hasDesktop {
		t.Fatalf("expected desktop toast, events=%v", rec.Events)
	}
	_ = cmd
}
```

If `tea.Batch` makes draining awkward, extract:

```go
func (m appModel) applyPhaseEnd(msg sessionPhaseEndMsg) (appModel, tea.Cmd) {
	m.hub.notice = fmt.Sprintf("%s — %s", msg.Title, msg.Body)
	m.hub.status = m.hub.notice
	m.hub.alerting = true
	m.hub.errMsg = ""
	_ = m.session.notifier.Desktop(msg.Title, msg.Body)
	return m, tea.Batch(hubAudioAlert(), scheduleAlertTick(), scheduleTick())
}
```

And store `notifier` on session from `newSessionModel` (already does). For app-level tests, pass `rec` into `newSessionModel`.

Also handle zero-remaining tick on hub by routing through session update:

```go
// inside appModel.Update, before page switch — or in pageHub branch:
```

Concrete hub+session tick logic in Step 3.

- [ ] **Step 2: Run — expect FAIL**

```powershell
go test ./internal/tui/ -run "TestHubTickAdvancesBackgroundPomodoro|TestPhaseEndOnHubSetsNoticeAndDesktop" -count=1
```

- [ ] **Step 3: Implement tick routing + applyPhaseEnd**

In `app.go` `Update`, handle `sessionPhaseEndMsg` at the top level:

```go
case sessionPhaseEndMsg:
	return m.applyPhaseEnd(msg)
```

```go
func (m appModel) applyPhaseEnd(msg sessionPhaseEndMsg) (tea.Model, tea.Cmd) {
	m.hub.notice = fmt.Sprintf("%s — %s", msg.Title, msg.Body)
	m.hub.status = m.hub.notice
	m.hub.alerting = true
	m.hub.errMsg = ""
	n := m.session.notifier
	if n == nil {
		n = notify.Default{}
	}
	_ = n.Desktop(msg.Title, msg.Body)
	m.syncHubPomodoro()
	return m, tea.Batch(hubAudioAlert(), scheduleAlertTick(), scheduleTick())
}
```

For hub page ticks, replace the simple hub-only update:

```go
case pageHub:
	if m.session.live() {
		// Advance pomodoro even while looking at the hub.
		sm, scmd := m.session.Update(msg)
		m.session = sm.(sessionModel)
		hm, hcmd := m.hub.Update(msg)
		m.hub = hm.(hubModel)
		m.syncHubPomodoro()
		return m, tea.Batch(scmd, hcmd)
	}
	mod, cmd := m.hub.Update(msg)
	m.hub = mod.(hubModel)
	return m, cmd
```

Important: when `msg` is `tickMsg` and session finishes a phase, `scmd` may be a `Batch` including `sessionPhaseEndMsg`. Bubble Tea executes cmds asynchronously and delivers msgs back to `Update` — that is fine in production. In unit tests, invoke `cmd()` and feed returned msgs into `Update` (or test `applyPhaseEnd` directly as above).

When `pageSession`, keep existing session update, but phase-end cmds still bubble to app on next delivery — **also** handle synchronously if session returns `sessionPhaseEndMsg` wrapped in Batch. Production Bubble Tea handles this. For session page Esc while alerting, see Task 5.

Ensure `session.Update(tickMsg)` when `remaining == 0` still calls `finishPhase` (today it decrements then checks `<= 0`). For hub-driven tick with `remaining == time.Second`, one tick ends the phase. Good.

**Paused on hub:** Spec says pause only applies on session page; while minimized, time continues. So when advancing from hub, clear pause or ignore pause:

In hub-driven path only:

```go
if m.session.live() && !m.session.waitingEnter {
	m.session.paused = false // background always runs
}
```

Or in `session.Update` add a flag — simpler to force `paused = false` when minimizing:

```go
case sessionMinimizeMsg:
	m.session.paused = false
	// ...
```

- [ ] **Step 4: Run — expect PASS**

```powershell
go test ./internal/tui/ -run "TestHubTickAdvancesBackgroundPomodoro|TestPhaseEndOnHubSetsNoticeAndDesktop|TestFinishPhaseDoesNotCallBanner" -count=1
```

- [ ] **Step 5: Commit**

```powershell
git add internal/tui/app.go internal/tui/app_test.go internal/tui/session.go
git commit -m "feat(tui): advance pomodoro on hub with notice and desktop alert"
```

---

### Task 5: Esc priority, waitingEnter on hub, footers

**Files:**
- Modify: `internal/tui/session.go` (Esc dismisses alarm first — needs hub alerting visible; handle in app)
- Modify: `internal/tui/app.go`
- Modify: `internal/tui/hub.go` (Enter advances when `PomodoroWaiting`)
- Test: `internal/tui/app_test.go`, `internal/tui/hub_test.go`

- [ ] **Step 1: Write failing tests**

```go
func TestSessionEscDismissesAlarmBeforeMinimize(t *testing.T) {
	m := initialModel(Dependencies{Root: t.TempDir(), Now: time.Now})
	m.session = newSessionModel(pomodoro.Config{
		Focus: time.Minute, Break: time.Minute, Long: time.Minute, Cycles: 1, Auto: true,
	}, 80, 24, nil)
	m.page = pageSession
	m.hub.alerting = true
	m.hub.notice = "Focus complete — Focus session 1 finished"
	mod, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = mod.(appModel)
	if m.hub.alerting || m.hub.notice != "" {
		t.Fatal("expected alarm cleared")
	}
	if m.page != pageSession {
		t.Fatal("should stay on session when dismissing alarm")
	}
}

func TestHubEnterAdvancesWaitingPomodoro(t *testing.T) {
	m := initialModel(Dependencies{Root: t.TempDir(), Now: time.Now})
	cfg := pomodoro.Config{Focus: time.Minute, Break: time.Minute, Long: time.Minute, Cycles: 2, Auto: false}
	m.session = newSessionModel(cfg, 80, 24, nil)
	m.session.waitingEnter = true
	m.session.status = "Press Enter or n to continue"
	m.page = pageHub
	m.syncHubPomodoro()
	idx := m.session.index
	mod, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = mod.(appModel)
	if m.session.waitingEnter {
		t.Fatal("still waiting")
	}
	if m.session.index != idx+1 {
		t.Fatalf("index=%d want %d", m.session.index, idx+1)
	}
}
```

- [ ] **Step 2: Run — expect FAIL**

```powershell
go test ./internal/tui/ -run "TestSessionEscDismissesAlarmBeforeMinimize|TestHubEnterAdvancesWaitingPomodoro" -count=1
```

- [ ] **Step 3: Implement**

In `app.go`, intercept Esc before delegating to session when on session page:

```go
case pageSession:
	if key, ok := msg.(tea.KeyMsg); ok {
		ks := key.String()
		if ks == "esc" || ks == "b" {
			if m.hub.notice != "" || m.hub.alerting {
				m.hub.notice = ""
				m.hub.alerting = false
				notify.StopAlert()
				return m, nil
			}
			return m.Update(sessionMinimizeMsg{})
		}
	}
	mod, cmd := m.session.Update(msg)
	m.session = mod.(sessionModel)
	m.syncHubPomodoro()
	return m, cmd
```

Remove duplicate Esc handling from session **or** leave it but unreachable when app intercepts — prefer remove Esc/b from session `Update` so minimize only goes through app (avoids double messages). Keep `q` on session.

Hub Enter when waiting — in `app.go` pageHub key intercept **or** in `hub.startSelected` / Enter handler. Cleanest in app before hub update:

```go
case pageHub:
	if key, ok := msg.(tea.KeyMsg); ok && key.String() == "enter" {
		if m.session.live() && m.session.waitingEnter {
			sm, scmd := m.session.advanceAfterWait()
			m.session = sm.(sessionModel)
			m.syncHubPomodoro()
			m.hub.status = "Next phase started"
			return m, tea.Batch(scmd, scheduleTick())
		}
	}
	// existing hub (+ background session tick) logic
```

Update footers:

- Session: `space pause  n skip/next  esc/b hub  q quit` (clarify hub minimize) — e.g. `esc alarm/hub  b hub  q quit`
- Hub: mention Open/Stop when relevant — existing footer can add `open/stop via Actions`

- [ ] **Step 4: Run — expect PASS**

```powershell
go test ./internal/tui/ -run "TestSessionEscDismissesAlarmBeforeMinimize|TestHubEnterAdvancesWaitingPomodoro|TestEscStopsAlarm" -count=1
```

- [ ] **Step 5: Commit**

```powershell
git add internal/tui/app.go internal/tui/app_test.go internal/tui/session.go internal/tui/hub.go
git commit -m "feat(tui): esc dismisses pomodoro alarm; hub Enter advances wait"
```

---

### Task 6: Full regression + manual checklist

**Files:** none new; run suite

- [ ] **Step 1: Run all package tests**

```powershell
go test ./...
```

Expected: all PASS. Fix any breakage in `TestHubActionBarInView`, session pause tests, or form launch paths that assumed Esc aborted.

- [ ] **Step 2: Manual checklist (interactive TTY)**

1. `go run ./cmd/clocky` → New Pomodoro with short focus (e.g. 5s), Auto=y.
2. Confirm ASCII clock ticks; wait for phase end → hear beep, see notice after returning to hub (or notice set while still in session then Esc once clears alarm).
3. During focus, Esc → hub shows `● Pomodoro … left`; remaining decreases.
4. Open Pomodoro → same remaining; Stop Pomodoro → idle.
5. Start pomodoro, try New Pomodoro again → error.
6. Auto=n: phase end on hub → waiting; Enter advances; alarm Esc-dismissible.
7. Confirm terminal is **not** corrupted (no `*** Focus complete ***` junk over the face).

- [ ] **Step 3: Commit any fixes**

```powershell
git add -A
git commit -m "test(tui): stabilize pomodoro background regression fixes"
```

(Skip empty commit if clean.)

---

## Spec coverage self-check

| Spec requirement | Task |
|------------------|------|
| No stdout Banner on TUI phase end | Task 1 |
| Notice + repeating Alert + Desktop | Task 4 |
| Esc minimizes (keeps session) | Task 3 |
| Hub Active shows phase/remaining | Task 2 |
| Open / Stop buttons | Task 2–3 |
| Hub tick advances session | Task 4 |
| Respect Auto / waiting on hub | Task 4–5 |
| Refuse second start | Task 3 |
| Esc dismisses alarm first | Task 5 |
| Pause only on session page; hub continues | Task 4 (clear pause on minimize) |
| Quit process ends pomodoro (no disk) | unchanged (no persistence added) |
| CLI `notify.All` unchanged | no CLI edits |

## Out of scope (do not implement)

- `pomodoro.json` / survive process exit
- Detached worker
- Changing CLI pomodoro banner behavior
