# Clocky Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a cross-platform Go CLI (`clocky`) with pomodoro (ASCII analog clock), background timer + notifications, stopwatch toggle, and named presets.

**Architecture:** Single Go module with `cmd/clocky` entrypoint and focused `internal/` packages (duration, state, presets, notify, clockface, stopwatch, timer, pomodoro). Persistent state under `~/.clocky/`. Pomodoro runs foreground; timer detaches a background worker process.

**Tech Stack:** Go 1.22+, standard library first; native notifications via thin OS wrappers (`osascript`, `notify-send`, PowerShell toast) behind a `Notifier` interface for tests.

**Spec:** `docs/superpowers/specs/2026-08-29-clocky-design.md`

---

## File map

| Path | Responsibility |
|------|----------------|
| `go.mod` | Module `github.com/luisf/clocky` (adjust module path if repo differs) |
| `cmd/clocky/main.go` | `os.Exit(run(os.Args[1:]))` thin wrapper |
| `internal/cli/cli.go` | Subcommand routing, flag parsing, help text |
| `internal/duration/duration.go` | Flexible `H:M:S` parse + `Format` |
| `internal/duration/duration_test.go` | Parse matrix tests |
| `internal/state/dir.go` | Resolve `~/.clocky`, ensure dir exists |
| `internal/state/jsonfile.go` | Atomic read/write JSON helpers |
| `internal/presets/presets.go` | Load/save/list/remove named pomodoro & timer presets |
| `internal/presets/presets_test.go` | Preset CRUD tests (temp dir) |
| `internal/notify/notify.go` | `Notifier` interface + `Beep`, `Banner`, `Desktop` |
| `internal/notify/desktop_*.go` | OS-specific desktop notify (build tags) |
| `internal/clockface/clockface.go` | ASCII analog face renderer |
| `internal/clockface/clockface_test.go` | Hand mapping / non-empty render tests |
| `internal/stopwatch/stopwatch.go` | Toggle start/stop via `stopwatch.json` |
| `internal/stopwatch/stopwatch_test.go` | Start/stop transitions |
| `internal/timer/timer.go` | Start background, stop, status, worker entry |
| `internal/timer/timer_test.go` | Duration/deadline helpers (no long sleeps) |
| `internal/pomodoro/pomodoro.go` | Phase machine + loop with injectable clock/notifier |
| `internal/pomodoro/pomodoro_test.go` | Cycle sequence tests |
| `README.md` | Already exists — update module path if needed |

---

### Task 0: Install Go and scaffold module

**Files:**
- Create: `go.mod`
- Create: `cmd/clocky/main.go`
- Create: `internal/cli/cli.go`

- [ ] **Step 1: Install Go 1.22+ on Windows if missing**

```powershell
winget install --id GoLang.Go -e --accept-package-agreements --accept-source-agreements
$env:Path = [System.Environment]::GetEnvironmentVariable("Path","Machine") + ";" + [System.Environment]::GetEnvironmentVariable("Path","User")
go version
```

Expected: `go version go1.22` or newer.

- [ ] **Step 2: Initialize module and hello CLI**

```powershell
cd C:\Users\luisf\orca\projects\Clocky
go mod init github.com/luisf/clocky
```

Create `cmd/clocky/main.go`:

```go
package main

import (
	"fmt"
	"os"

	"github.com/luisf/clocky/internal/cli"
)

func main() {
	if err := cli.Run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
```

Create `internal/cli/cli.go`:

```go
package cli

import "fmt"

func Run(args []string) error {
	if len(args) == 0 {
		printHelp()
		return nil
	}
	switch args[0] {
	case "help", "-h", "--help":
		printHelp()
		return nil
	default:
		return fmt.Errorf("unknown command %q\nrun: clocky help", args[0])
	}
}

func printHelp() {
	fmt.Print(`Clocky — terminal time manager

Usage:
  clocky pomodoro [name] [flags]
  clocky timer <duration|name>
  clocky timer --stop
  clocky stopwatch
  clocky add pomodoro --name <name> [flags]
  clocky add timer --name <name> <duration>
  clocky list [pomodoro|timer]
  clocky remove <pomodoro|timer> <name>
  clocky status
  clocky help
`)
}
```

- [ ] **Step 3: Build**

```powershell
go build -o clocky.exe ./cmd/clocky
.\clocky.exe help
```

Expected: help text printed; exit 0.

- [ ] **Step 4: Commit**

```bash
git add go.mod go.sum cmd/clocky/main.go internal/cli/cli.go
git commit -m "chore: scaffold clocky Go module and CLI stub"
```

---

### Task 1: Duration parser

**Files:**
- Create: `internal/duration/duration.go`
- Create: `internal/duration/duration_test.go`

- [ ] **Step 1: Write failing tests**

```go
package duration

import (
	"testing"
	"time"
)

func TestParse(t *testing.T) {
	cases := []struct {
		in   string
		want time.Duration
	}{
		{"1:30:00", time.Hour + 30*time.Minute},
		{"1:30:", time.Hour + 30*time.Minute},
		{"1::30", time.Hour + 30*time.Second},
		{":5:30", 5*time.Minute + 30*time.Second},
		{"24::", 24 * time.Hour},
		{":25:", 25 * time.Minute},
		{"::50", 50 * time.Second},
		{":1440:", 24 * time.Hour},
		{"::1440", 24 * time.Minute},
		{":90:00", 90 * time.Minute},
		{"90", 90 * time.Second},
	}
	for _, tc := range cases {
		got, err := Parse(tc.in)
		if err != nil {
			t.Fatalf("Parse(%q): %v", tc.in, err)
		}
		if got != tc.want {
			t.Fatalf("Parse(%q)=%v want %v", tc.in, got, tc.want)
		}
	}
}

func TestParseInvalid(t *testing.T) {
	for _, in := range []string{"", "abc", "1:2:3:4", "::", ":"} {
		if _, err := Parse(in); err == nil {
			t.Fatalf("Parse(%q) expected error", in)
		}
	}
}

func TestFormat(t *testing.T) {
	if got := Format(time.Hour + 2*time.Minute + 3*time.Second); got != "01:02:03" {
		t.Fatalf("got %q", got)
	}
	if got := Format(24 * time.Hour); got != "24:00:00" {
		t.Fatalf("got %q", got)
	}
}
```

- [ ] **Step 2: Run tests — expect fail**

```powershell
go test ./internal/duration/ -v
```

Expected: FAIL (undefined Parse/Format).

- [ ] **Step 3: Implement**

```go
package duration

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

func Parse(s string) (time.Duration, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, fmt.Errorf("empty duration")
	}
	if !strings.Contains(s, ":") {
		sec, err := strconv.Atoi(s)
		if err != nil || sec < 0 {
			return 0, fmt.Errorf("invalid duration %q", s)
		}
		return time.Duration(sec) * time.Second, nil
	}
	parts := strings.Split(s, ":")
	if len(parts) != 3 {
		return 0, fmt.Errorf("invalid duration %q", s)
	}
	parsePart := func(p string) (int, error) {
		if p == "" {
			return 0, nil
		}
		n, err := strconv.Atoi(p)
		if err != nil || n < 0 {
			return 0, fmt.Errorf("invalid duration %q", s)
		}
		return n, nil
	}
	h, err := parsePart(parts[0])
	if err != nil {
		return 0, err
	}
	m, err := parsePart(parts[1])
	if err != nil {
		return 0, err
	}
	sec, err := parsePart(parts[2])
	if err != nil {
		return 0, err
	}
	if parts[0] == "" && parts[1] == "" && parts[2] == "" {
		return 0, fmt.Errorf("invalid duration %q", s)
	}
	// Disallow lone ":" forms with no digits already handled; "::" has three empty parts
	if h == 0 && m == 0 && sec == 0 && parts[0] == "" && parts[1] == "" && parts[2] == "" {
		return 0, fmt.Errorf("invalid duration %q", s)
	}
	total := time.Duration(h)*time.Hour + time.Duration(m)*time.Minute + time.Duration(sec)*time.Second
	if total == 0 && (parts[0] != "0" && parts[1] != "0" && parts[2] != "0") {
		// allow explicit zeros like 0:0:0? Spec timers of zero are useless — reject zero.
		return 0, fmt.Errorf("duration must be > 0")
	}
	if total <= 0 {
		return 0, fmt.Errorf("duration must be > 0")
	}
	return total, nil
}

func Format(d time.Duration) string {
	if d < 0 {
		d = -d
	}
	totalSec := int64(d / time.Second)
	h := totalSec / 3600
	m := (totalSec % 3600) / 60
	s := totalSec % 60
	return fmt.Sprintf("%02d:%02d:%02d", h, m, s)
}
```

Note: reject `0:0:0` and empty `::`. Keep `TestParseInvalid` covering `::` and `:`.

- [ ] **Step 4: Run tests — expect pass**

```powershell
go test ./internal/duration/ -v
```

- [ ] **Step 5: Commit**

```bash
git add internal/duration/
git commit -m "feat: add flexible H:M:S duration parser"
```

---

### Task 2: State directory + JSON helpers

**Files:**
- Create: `internal/state/dir.go`
- Create: `internal/state/jsonfile.go`
- Create: `internal/state/state_test.go`

- [ ] **Step 1: Write failing tests** using `t.TempDir()` and overriding base via `state.SetRootForTest(dir)` (or pass `root string` into helpers — prefer explicit `root` parameter to avoid globals).

API:

```go
package state

func EnsureDir(root string) error
func ReadJSON(path string, v any) error   // os.ErrNotExist if missing
func WriteJSON(path string, v any) error  // atomic write via temp + rename
func Path(root, name string) string
```

Test: write struct, read back; missing file returns `os.ErrNotExist`.

- [ ] **Step 2: Run fail → implement → pass → commit**

```bash
git commit -m "feat: add ~/.clocky state JSON helpers"
```

Default root helper:

```go
func DefaultRoot() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".clocky"), nil
}
```

---

### Task 3: Presets store

**Files:**
- Create: `internal/presets/presets.go`
- Create: `internal/presets/presets_test.go`

Types:

```go
type PomodoroPreset struct {
	Name   string `json:"name"`
	Focus  int    `json:"focus"`  // minutes
	Break  int    `json:"break"`
	Long   int    `json:"long"`
	Cycles int    `json:"cycles"`
	Auto   bool   `json:"auto"`
}

type TimerPreset struct {
	Name     string `json:"name"`
	Duration string `json:"duration"` // canonical Format(d) or original; store seconds as int64 preferred
	Seconds  int64  `json:"seconds"`
}

type Store struct {
	Pomodoros []PomodoroPreset `json:"pomodoros"`
	Timers    []TimerPreset    `json:"timers"`
}
```

API on `Store` loaded from `filepath.Join(root, "presets.json")`:

- `Load(root) (*Store, error)` — empty store if missing
- `Save(root) error`
- `UpsertPomodoro(p PomodoroPreset)`
- `UpsertTimer(t TimerPreset)`
- `GetPomodoro(name) (PomodoroPreset, bool)`
- `GetTimer(name) (TimerPreset, bool)`
- `RemovePomodoro(name) error`
- `RemoveTimer(name) error`

- [ ] **Step 1: Tests for upsert/get/overwrite/remove/list round-trip**
- [ ] **Step 2: Implement → pass → commit**

```bash
git commit -m "feat: add named pomodoro and timer presets"
```

---

### Task 4: Notify interface

**Files:**
- Create: `internal/notify/notify.go`
- Create: `internal/notify/desktop.go` (default stub calling OS helpers)
- Create: `internal/notify/desktop_windows.go`
- Create: `internal/notify/desktop_darwin.go`
- Create: `internal/notify/desktop_linux.go`
- Create: `internal/notify/notify_test.go`

```go
type Notifier interface {
	Beep() error
	Banner(title, body string) error
	Desktop(title, body string) error
}

type Default struct {
	Out io.Writer // for Banner; default os.Stdout
}

func (d Default) Beep() error {
	fmt.Fprint(d.out(), "\a")
	return nil
}

func (d Default) Banner(title, body string) error {
	fmt.Fprintf(d.out(), "\n*** %s ***\n%s\n\n", title, body)
	return nil
}
```

Desktop:
- Windows: `powershell -NoProfile -Command` with a short toast script, or write to stdout only if it fails
- Darwin: `osascript -e 'display notification "body" with title "title"'`
- Linux: `notify-send title body`

Tests: `RecordingNotifier` struct appending events; unit-test that `Default.Banner` writes expected text.

- [ ] Implement → test → commit

```bash
git commit -m "feat: add beep, banner, and native desktop notifications"
```

---

### Task 5: ASCII clock face

**Files:**
- Create: `internal/clockface/clockface.go`
- Create: `internal/clockface/clockface_test.go`

```go
// Render returns a multi-line ASCII analog clock for remaining duration
// within total (hands from progress), plus optional use of wall seconds.
func Render(remaining, total time.Duration, width int) string
```

Algorithm:
- Build ellipse/circle of width ~21 height ~11 using `o`/`.` for rim
- Map minute angle: `θ = 2π * (1 - remaining/total)` or based on wall clock seconds of remaining
- Spec: hands show countdown progress — use `elapsed/total` so hand moves as time passes
- Center `+`, hand cells `*` or `|`/`/`/`-`/`\`

Tests:
- `Render(25*time.Minute, 25*time.Minute, 21)` contains rim chars and is > 5 lines
- Different remaining values produce different strings

Fallback helper:

```go
func RenderCompact(phase string, cycle int, totalCycles int, remaining time.Duration) string
```

- [ ] TDD → commit

```bash
git commit -m "feat: add ASCII analog clock face renderer"
```

---

### Task 6: Stopwatch

**Files:**
- Create: `internal/stopwatch/stopwatch.go`
- Create: `internal/stopwatch/stopwatch_test.go`

```go
type State struct {
	StartedAt time.Time `json:"started_at"`
}

// Toggle returns ("started", 0, nil) or ("stopped", elapsed, nil)
func Toggle(root string, now time.Time) (action string, elapsed time.Duration, err error)
func Status(root string, now time.Time) (running bool, elapsed time.Duration, err error)
```

File: `stopwatch.json`. If missing/empty → start. If present → stop and delete file.

- [ ] TDD with temp root → commit

```bash
git commit -m "feat: add stopwatch start/stop toggle"
```

---

### Task 7: Pomodoro engine

**Files:**
- Create: `internal/pomodoro/pomodoro.go`
- Create: `internal/pomodoro/pomodoro_test.go`

```go
type Config struct {
	Focus  time.Duration
	Break  time.Duration
	Long   time.Duration
	Cycles int
	Auto   bool
}

type Hooks struct {
	Notify Notifier
	Sleep  func(d time.Duration) // default time.Sleep; tests use virtual
	Now    func() time.Time
	In     io.Reader  // Enter wait
	Out    io.Writer
	Width  int
}

func Run(cfg Config, h Hooks) error
```

Phase sequence for `Cycles=2`: Focus1 → Break → Focus2 → LongBreak → done.

Each phase:
1. Tick every 1s (or jump in tests via fake Sleep that still calls tick once)
2. Redraw clockface + `FOCUS`/`BREAK`/`LONG BREAK` + `i/n` + `MM:SS`
3. On end: notify beep+banner+desktop
4. If !Auto && more phases: wait for newline on `In`

For unit tests: use `Sleep` that no-ops and a shortened config with `Focus: time.Second` — or better, extract `PlanPhases(cfg) []Phase` and test the plan; separately test one tick render.

Minimum tests:
- `PlanPhases` returns correct order/lengths
- `Run` with Auto + Focus/Break/Long = 1ms + Cycles=1 calls notify at least twice

- [ ] TDD → commit

```bash
git commit -m "feat: add pomodoro phase engine with ASCII clock"
```

---

### Task 8: Background timer

**Files:**
- Create: `internal/timer/timer.go`
- Create: `internal/timer/timer_test.go`
- Modify: `cmd/clocky/main.go` or worker flag in cli

Design:
- `timer.json`: `{ "deadline": RFC3339, "pid": int, "label": string }`
- `Start(root, d, label, execPath, args)` writes state and starts child:
  - Child args: `clocky timer --worker`
  - Windows: `cmd.SysProcAttr` with detach / `CREATE_NEW_PROCESS_GROUP`
  - Unix: `Setpgid: true`
- Worker: sleep until deadline (loop 1s), then notify, clear state, exit
- `Stop(root)` kills PID, removes file
- `Status(root, now)` remaining time

Tests:
- Serialize deadline helpers
- `Stop` on missing → clear error message
- Integration optional: start worker with 1s in subprocess during Task 9

```bash
git commit -m "feat: add background timer with stop and status"
```

---

### Task 9: Wire CLI commands

**Files:**
- Modify: `internal/cli/cli.go`
- Create: `internal/cli/add.go`, `pomodoro.go`, `timer.go`, `stopwatch.go`, `list.go`, `status.go` (split if `cli.go` grows)

Implement routing:

| Args | Action |
|------|--------|
| `pomodoro ...` | parse flags + optional name → presets → `pomodoro.Run` |
| `timer --stop` | `timer.Stop` |
| `timer --worker` | `timer.Worker` (hidden) |
| `timer <dur\|name>` | resolve → `timer.Start` |
| `stopwatch` | `stopwatch.Toggle` print result |
| `add pomodoro --name N ...` | upsert preset |
| `add timer --name N <dur>` | upsert preset |
| `list [type]` | print presets |
| `remove <type> <name>` | remove |
| `status` | print stopwatch + timer |
| `help` | help |

Flag parsing: use `flag` package with a new `FlagSet` per subcommand.

Defaults for pomodoro: focus 25, break 5, long 15, cycles 4.

- [ ] Manual smoke:

```powershell
go build -o clocky.exe ./cmd/clocky
.\clocky.exe add pomodoro --name Focus --focus 1 --break 1 --long 1 --cycles 1 --auto
.\clocky.exe list
.\clocky.exe stopwatch
.\clocky.exe stopwatch
.\clocky.exe timer ::2
.\clocky.exe status
# wait ~2s for notification
.\clocky.exe timer --stop
```

- [ ] Commit

```bash
git commit -m "feat: wire all clocky CLI subcommands"
```

---

### Task 10: Polish, README path, full test

- [ ] `go test ./...` all green
- [ ] Fix Windows detach if timer child does not survive parent exit
- [ ] Update README clone/module path if known
- [ ] Final commit

```bash
git commit -m "chore: polish clocky MVP and verify tests"
```

---

## Spec coverage checklist

| Spec item | Task |
|-----------|------|
| Flexible duration + overflow | Task 1 |
| `~/.clocky` state | Task 2 |
| Presets add/list/remove + invoke by name | Tasks 3, 9 |
| Notify beep+banner+native | Task 4 |
| ASCII analog pomodoro | Tasks 5, 7 |
| Stopwatch toggle | Task 6 |
| Background timer + stop + status | Tasks 8, 9 |
| English CLI | Tasks 0, 9 |
| Unit tests duration/presets/cycles | Tasks 1, 3, 7 |

## Self-review notes

- No Portuguese command names; command is `stopwatch`.
- Zero duration rejected by parser.
- Timer worker invoked via hidden `--worker` flag (same binary).
- Module path `github.com/luisf/clocky` is provisional — change if the GitHub repo name/owner differs.
