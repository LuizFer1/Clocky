# Clocky Terminal UI Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Ship an interactive Bubble Tea dashboard so `clocky` (no args, TTY) manages live timer/stopwatch, pomodoro sessions with pause, and full preset CRUD — sharing `~/.clocky/` with the existing CLI.

**Architecture:** New `internal/tui` Bubble Tea app (hub → session → forms). Domain packages unchanged as source of truth. Empty CLI args on a TTY call `tui.Run()`; non-TTY keeps printing help.

**Tech Stack:** Go, Bubble Tea, Lip Gloss, Bubbles; existing `timer` / `stopwatch` / `presets` / `pomodoro` / `clockface` / `notify` / `duration` / `state`.

**Spec:** `docs/superpowers/specs/2026-08-30-clocky-terminal-ui-design.md`

---

## File map

| Path | Responsibility |
|------|----------------|
| `internal/tui/run.go` | `Run()` entry, TTY check export used by CLI |
| `internal/tui/app.go` | Root model: view routing (hub / session / form / confirm) |
| `internal/tui/styles.go` | Lip Gloss styles (borders, titles, selected, error, muted) |
| `internal/tui/keys.go` | Key matching helpers |
| `internal/tui/tick.go` | `tickMsg` + `tea.Tick` every 1s |
| `internal/tui/hub.go` | Hub view: Active + Presets list + key handlers |
| `internal/tui/session.go` | Pomodoro session model (pause, skip, clockface, notify) |
| `internal/tui/form.go` | New/edit preset forms (pomodoro + timer) |
| `internal/tui/confirm.go` | Delete confirmation |
| `internal/tui/active.go` | Load Active snapshot from timer/stopwatch Status |
| `internal/tui/presets_list.go` | Unified preset items for Bubbles list |
| `internal/tui/*_test.go` | Model Update tests (pause, quit, tick, CRUD reduce) |
| `internal/cli/cli.go` | `len(args)==0` → TUI when TTY; else help; update help text |
| `internal/cli/tty.go` | `isInteractive() bool` (stdout terminal) |
| `README.md` | Document bare `clocky` opens TUI |

---

### Task 1: Dependencies, TTY gate, CLI entry

**Files:**
- Modify: `go.mod` / `go.sum` (add charmbracelet deps)
- Create: `internal/cli/tty.go`
- Create: `internal/tui/run.go` (stub `Run` returning nil initially is OK only until Task 2 — prefer minimal real Program)
- Modify: `internal/cli/cli.go`
- Test: `internal/cli/tty_test.go` (optional) / rely on build

- [ ] **Step 1: Add modules**

```powershell
go get github.com/charmbracelet/bubbletea@latest
go get github.com/charmbracelet/lipgloss@latest
go get github.com/charmbracelet/bubbles@latest
go mod tidy
```

Expected: `go.mod` lists the three modules.

- [ ] **Step 2: TTY helper**

Create `internal/cli/tty.go`:

```go
package cli

import (
	"os"

	"golang.org/x/term"
)

func isInteractive() bool {
	return term.IsTerminal(int(os.Stdout.Fd())) && term.IsTerminal(int(os.Stdin.Fd()))
}
```

- [ ] **Step 3: Wire empty args**

In `internal/cli/cli.go` `Run`:

```go
if len(args) == 0 {
	if isInteractive() {
		return tui.Run()
	}
	printHelp()
	return nil
}
```

Update `printHelp` to include:

```text
  clocky                 Open the Terminal UI (interactive terminal)
```

- [ ] **Step 4: Stub `tui.Run`**

```go
package tui

func Run() error {
	return fmt.Errorf("tui not implemented")
}
```

Replace in Task 2 with real Program.

- [ ] **Step 5: Commit**

```bash
git add go.mod go.sum internal/cli/cli.go internal/cli/tty.go internal/tui/run.go
git commit -m "feat: gate interactive TUI entry from bare clocky"
```

---

### Task 2: App shell, styles, tick, empty hub render

**Files:**
- Create: `internal/tui/styles.go`, `tick.go`, `app.go`, `hub.go` (minimal View)
- Modify: `internal/tui/run.go`
- Test: `internal/tui/app_test.go`

- [ ] **Step 1: Failing test — quit key**

```go
func TestAppQuit(t *testing.T) {
	m := newAppModel(testDeps(t))
	m2, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	if cmd == nil {
		t.Fatal("expected quit cmd")
	}
	// Execute cmd batch; look for tea.Quit
	_ = m2
}
```

Use a helper that runs `tea.Batch` cmds or check `Update` returns `tea.Quit` via inspecting with a small wrapper: prefer asserting model `quitting` flag set OR that cmd equals `tea.Quit` (compare via running in `tea.NewProgram` with fake input — simplest: set `m.quitting` and `Init`/`Update` return `tea.Quit` when quitting).

Simpler approach used in Charm tests:

```go
func TestHubQuitSetsCommand(t *testing.T) {
	m := initialModel(Dependencies{Root: t.TempDir(), Now: time.Now})
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})
	if cmd == nil {
		t.Fatal("expected cmd")
	}
}
```

- [ ] **Step 2: Implement shell**

`Dependencies` struct:

```go
type Dependencies struct {
	Root string
	Now  func() time.Time
	Exe  func() (string, error) // for timer.Start
}
```

`model` fields: `deps`, `width`, `height`, `page` (`pageHub|pageSession|pageForm|pageConfirm`), `hub hubModel`, `session sessionModel`, `form formModel`, `errMsg string`.

`tickMsg struct{}` + `func scheduleTick() tea.Cmd` → `tea.Tick(time.Second, func(t time.Time) tea.Msg { return tickMsg{} })`.

`Run()`:

```go
func Run() error {
	root, err := state.DefaultRoot()
	// EnsureDir
	m := initialModel(Dependencies{
		Root: root,
		Now:  time.Now,
		Exe:  os.Executable,
	})
	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err = p.Run()
	return err
}
```

Hub `View` at least prints titles `Active` and `Presets` and footer with `q quit`.

- [ ] **Step 3: Tests pass + commit**

```bash
go test ./internal/tui/ -count=1
git add internal/tui
git commit -m "feat: add Bubble Tea app shell and hub chrome"
```

---

### Task 3: Active panel — live timer/stopwatch + actions

**Files:**
- Create: `internal/tui/active.go`
- Modify: `internal/tui/hub.go`, `app.go`
- Test: `internal/tui/active_test.go`

- [ ] **Step 1: Failing tests**

```go
func TestRefreshActiveReadsStopwatch(t *testing.T) {
	root := t.TempDir()
	_, _, err := stopwatch.Toggle(root, time.Unix(1000, 0))
	if err != nil { t.Fatal(err) }
	a := refreshActive(root, time.Unix(1010, 0))
	if !a.StopwatchRunning || a.StopwatchElapsed != 10*time.Second {
		t.Fatalf("%+v", a)
	}
}

func TestToggleStopwatchKey(t *testing.T) {
	// hub Update on 't' starts stopwatch file under deps.Root
}
```

- [ ] **Step 2: Implement**

```go
type activeSnapshot struct {
	TimerActive     bool
	TimerRemaining  time.Duration
	TimerLabel      string
	StopwatchRunning bool
	StopwatchElapsed time.Duration
}

func refreshActive(root string, now time.Time) activeSnapshot {
	// timer.Status + stopwatch.Status; ignore errors → zero
}
```

Keys on hub:
- `s` → `timer.Stop(root)` then refresh; set `errMsg` on failure
- `t` → `stopwatch.Toggle(root, now)` then refresh
- `p` → switch to session with default config (25/5/15/4, auto false) — session body in Task 5; for now may stub page switch

On `tickMsg`, refresh active and `scheduleTick()`.

Render Active panel with lipgloss border; show `remaining` / `elapsed` via `duration.Format`.

- [ ] **Step 3: Commit**

```bash
git commit -m "feat: live Active panel for timer and stopwatch"
```

---

### Task 4: Unified presets list + start

**Files:**
- Create: `internal/tui/presets_list.go`
- Modify: `hub.go`
- Test: `internal/tui/presets_list_test.go`

- [ ] **Step 1: Failing test — build items**

```go
func TestBuildPresetItems(t *testing.T) {
	s := &presets.Store{
		Pomodoros: []presets.PomodoroPreset{{Name: "Focus", Focus: 40, Break: 10, Long: 15, Cycles: 4}},
		Timers:    []presets.TimerPreset{{Name: "Break", Seconds: 300}},
	}
	items := buildPresetItems(s)
	if len(items) != 2 { t.Fatal(len(items)) }
}
```

- [ ] **Step 2: Implement list**

Use `github.com/charmbracelet/bubbles/list` or a simple index + slice (prefer simple cursor for fewer deps surface):

```go
type presetItem struct {
	Kind string // "pomodoro" | "timer"
	Name string
	Summary string
	// embed original preset fields needed to start
	Pomo *presets.PomodoroPreset
	Tim  *presets.TimerPreset
}
```

Keys:
- `↑`/`↓` or `k`/`j` move cursor
- `Enter` start:
  - pomodoro → open session with that config
  - timer → `timer.Start(root, d, name, exe, []string{"timer","--worker"})`; on "already running" set errMsg

Reload presets on hub enter and after CRUD.

- [ ] **Step 3: Commit**

```bash
git commit -m "feat: hub presets list with start"
```

---

### Task 5: Pomodoro session view (clock, pause, skip, notify)

**Files:**
- Create: `internal/tui/session.go`, `session_test.go`
- Modify: `app.go`, `hub.go`

- [ ] **Step 1: Failing tests**

```go
func TestSessionPauseFreezesRemaining(t *testing.T) {
	s := newSession(pomodoro.Config{Focus: 2 * time.Second, Break: time.Second, Long: time.Second, Cycles: 1, Auto: true}, Dependencies{Now: mono(0)})
	s2, _ := s.Update(tickMsg{}) // remaining -= 1s if not paused
	s3, _ := s2.(sessionModel).Update(tea.KeyMsg{Type: tea.KeySpace})
	before := s3.(sessionModel).remaining
	s4, _ := s3.(sessionModel).Update(tickMsg{})
	if s4.(sessionModel).remaining != before {
		t.Fatal("paused should not decrease")
	}
}
```

- [ ] **Step 2: Implement session**

State: `phases []pomodoro.Phase`, `index int`, `remaining`, `total`, `paused bool`, `waitingEnter bool` (when !Auto between phases), `cfg`, `width`.

On tick (if !paused && !waitingEnter): decrement 1s; at 0 → `notify.All`, then either wait Enter or auto-advance; after last phase → `sessionDoneMsg` → hub.

Keys:
- `Space` toggle pause (ignored while waitingEnter)
- `n` skip: jump to phase end notify path
- `esc`/`b` → `sessionAbortMsg` → hub
- View: `clockface.Render(remaining, total, width)` + compact + block progress + `PAUSED` + footer

Progress bar helper:

```go
func progressBar(remaining, total time.Duration, width int) string {
	if total <= 0 || width <= 0 { return "" }
	done := 1 - float64(remaining)/float64(total)
	filled := int(done * float64(width))
	// █ and ░
}
```

- [ ] **Step 3: Commit**

```bash
git commit -m "feat: pomodoro session view with pause and ASCII clock"
```

---

### Task 6: Preset CRUD forms + delete confirm

**Files:**
- Create: `internal/tui/form.go`, `confirm.go`, tests
- Modify: `hub.go`, `app.go`

- [ ] **Step 1: Tests**

```go
func TestUpsertPomodoroFromForm(t *testing.T) {
	root := t.TempDir()
	err := savePomodoroPreset(root, presets.PomodoroPreset{Name: "X", Focus: 25, Break: 5, Long: 15, Cycles: 4})
	if err != nil { t.Fatal(err) }
	st, _ := presets.Load(root)
	if _, ok := st.GetPomodoro("X"); !ok { t.Fatal("missing") }
}
```

- [ ] **Step 2: Forms**

`n` → type picker (pomodoro/timer) → form with bubbles/textinput fields.  
`e` → edit selected.  
`d` → confirm (`y`/`n`).

Pomodoro fields: name, focus, break, long, cycles, auto (`true`/`false` or `y`/`n`).  
Timer fields: name, duration string → `duration.Parse`.

On save: `Load`, `Upsert*`, `Save`, return hub + reload list. Inline `errMsg` on validation failure (empty name, cycles < 1, parse error).

- [ ] **Step 3: Commit**

```bash
git commit -m "feat: TUI preset create edit delete"
```

---

### Task 7: Polish — help, README, non-TTY, resize

**Files:**
- Modify: `internal/cli/cli.go`, `README.md`, `styles.go` / hub layout for narrow width
- Test: ensure `go test ./...`

- [ ] **Step 1: README** — document `clocky` opens TUI; keep CLI examples.

- [ ] **Step 2: Handle `tea.WindowSizeMsg`** — store width/height; pass width into clockface.

- [ ] **Step 3: Full test suite**

```powershell
go test ./... -count=1
```

Expected: all pass.

- [ ] **Step 4: Manual smoke (dev binary)**

```powershell
.\scripts\build-dev.ps1
# In a real terminal: .\clockyDEV.exe
# Verify hub, start stopwatch, quit, status still running, reopen
```

- [ ] **Step 5: Final commit**

```bash
git commit -m "docs: document Terminal UI entry and hub usage"
```

---

## Spec coverage checklist

| Spec item | Task |
|-----------|------|
| Bare `clocky` → TUI on TTY | 1 |
| Non-TTY → help | 1 |
| Hub Active + Presets | 2–4 |
| Live tick timer/stopwatch | 3 |
| Continue after quit | 3 (domain persistence) |
| Start timer/stopwatch from hub | 3–4 |
| Preset CRUD | 6 |
| Pomodoro ASCII + pause | 5 |
| Skip / back to hub | 5 |
| Rich style no emoji | 2, 7 |
| Tests | 2–6 |
| README | 7 |

## Notes for implementers

- Prefer thin wrappers over duplicating detach/notify logic.
- Do not persist pomodoro mid-session.
- One timer at a time — surface existing `timer already running` errors in the hub status line.
- Keep English UI strings.
`)