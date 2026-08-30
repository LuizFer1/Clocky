# Clocky — Terminal UI Design

**Date:** 2026-08-30  
**Status:** Approved for planning  
**Platform:** Windows, macOS, Linux (interactive terminal)  
**Branch:** `feat/terminal-ui`

## Goal

Add a modern fullscreen Terminal UI so Clocky is easier to use day to day: manage saved pomodoros and timers, see and control in-progress timer/stopwatch with live ticks, and run pomodoro sessions with the existing ASCII analog clock — while keeping the current CLI subcommands for scripting and quick one-shots.

**Locale:** UI labels, help, and errors remain **English** (same as CLI).

## Decisions (from brainstorming)

| Topic | Choice |
|-------|--------|
| Entry point | `clocky` with no args opens the TUI; CLI subcommands stay |
| Home screen | Status + actions hub; ASCII clock appears in pomodoro session view |
| Visual richness | Rich terminal chrome without emoji: box-drawing, block elements (`░▒▓█`), ANSI colors, panels |
| Active timer/stopwatch | Live 1s tick in the hub; start/stop/toggle without leaving the TUI |
| Preset management | Full CRUD in the TUI (create, edit, delete, start); CLI add/list/remove remain |
| Close TUI while running | Timer and stopwatch **continue in background**; hub resumes from `~/.clocky/` on reopen |
| Implementation | **Bubble Tea + Lip Gloss** (+ Bubbles for lists/inputs) |
| Pomodoro pause | **In MVP:** Space toggles pause/resume; face and MM:SS freeze while paused |

## Approach

**A — Bubble Tea dashboard over shared domain packages.**

New `internal/tui` owns the interactive app. Existing `pomodoro`, `timer`, `stopwatch`, `presets`, `clockface`, `state`, `notify`, and `duration` packages stay the source of truth. No second state store.

## Architecture

```text
clocky
├── cmd/clocky/
└── internal/
    ├── cli/          # subcommands; len(args)==0 → tui.Run() when TTY
    ├── tui/          # Bubble Tea models: app shell, hub, session, forms
    ├── pomodoro/
    ├── timer/
    ├── stopwatch/
    ├── presets/
    ├── clockface/
    ├── duration/
    ├── state/
    ├── notify/
    └── termui/       # keep for CLI pomodoro path / ANSI enable helpers as needed
```

### Dependencies

- `github.com/charmbracelet/bubbletea`
- `github.com/charmbracelet/lipgloss`
- `github.com/charmbracelet/bubbles` (list, textinput, and related)

### State layout (`~/.clocky/`) — unchanged files

| Path | Purpose |
|------|---------|
| `presets.json` | Named pomodoro and timer presets (CRUD from TUI and CLI) |
| `timer.json` | Active background timer (deadline, PID, label) |
| `stopwatch.json` | Active stopwatch (`started_at`) or cleared when stopped |

Pomodoro session progress is **in-memory only** inside the TUI session view (closing the TUI ends that session; it does not auto-resume). Timer/stopwatch persist as today.

## Screens and components

### Hub (default view)

Panels:

1. **Active** — live status for timer (remaining), stopwatch (elapsed), and pomodoro (idle or “open session”). Actions: stop timer, toggle stopwatch, start pomodoro (default or via preset selection).
2. **Presets** — unified list of pomodoro and timer presets (type, name, summary). Select + Enter starts; New / Edit / Delete for CRUD.
3. **Footer / help** — keybindings.

Sketch (informative, not pixel-perfect):

```text
┌─ Clocky ──────────────────────────────────────────────┐
│ Active                                                │
│  ● Timer      Break     12:34 remaining    [S]top     │
│  ● Stopwatch            00:05:12 elapsed   [T]oggle   │
│  ○ Pomodoro             idle               [P] start  │
├─ Presets ────────────────────────── [N]ew [E]dit [D] ─┤
│  ▶ Focus    pomo   40/10/15 ×4 auto                   │
│    Break    timer  :05:                               │
├─ ↑↓ select  ↵ start  n/e/d  q quit ───────────────────┤
└───────────────────────────────────────────────────────┘
```

### Pomodoro session view

- Existing `clockface.Render` / `RenderCompact` for the ASCII analog face (remaining-time hands).
- Phase / cycle / MM:SS plus a block progress bar (`░▒▓█`).
- **Space:** pause / resume (paused label; ticker does not advance remaining).
- Skip phase / return to hub / quit app (quit does not kill background timer/stopwatch).
- Phase end: beep + ASCII banner + native OS notification via `notify` (same as CLI).

### CRUD forms

Bubble Tea views (or modal-style models) using Bubbles inputs:

| Form | Fields |
|------|--------|
| Pomodoro preset | name, focus, break, long, cycles, auto |
| Timer preset | name, duration (flexible `H:M:S` via `duration` package) |

Validation mirrors CLI rules. Overwrite existing name: allow with confirmation message (same MVP policy as CLI).

## Data flow

```text
Hub tick (1s)
  → read timer.json / stopwatch.json (or cached deadlines)
  → recompute remaining / elapsed
  → redraw Active panel

Start timer from TUI
  → same detach + timer.json path as `clocky timer`
  → appears in Active with live remaining

Stop timer
  → same as `clocky timer --stop`

Toggle stopwatch
  → same start/stop semantics as `clocky stopwatch`
  → when running, hub shows live elapsed from started_at

Start pomodoro (preset or defaults)
  → switch to session model; run phases in-process
  → on leave session → hub (pomodoro not persisted)

CRUD presets
  → presets.Load/Save on presets.json
  → list refresh
```

**Concurrency with CLI:** one active timer at a time. If CLI already started a timer, the hub shows it and refuses a second start until stop. Stale PID → clear state and show idle (reuse existing alive checks).

## Keybindings (MVP)

| Key | Context | Action |
|-----|---------|--------|
| `↑` / `↓` | Hub presets | Move selection |
| `Enter` | Hub | Start selected preset |
| `n` | Hub | New preset (type picker then form) |
| `e` | Hub | Edit selected preset |
| `d` | Hub | Delete selected (confirm) |
| `s` | Hub | Stop active timer |
| `t` | Hub | Toggle stopwatch |
| `p` | Hub | Start default pomodoro session |
| `Space` | Session | Pause / resume |
| `n` | Session | Skip / advance to next phase (after notify rules; respects `--auto`-equivalent session flag from preset/defaults) |
| `esc` / `b` | Session | End session and return to hub |
| `q` / Ctrl+C | Global | Quit TUI; leave timer/stopwatch running |

Session footer must show these bindings.

## Entry and non-TTY behavior

- Interactive TTY + no args → `tui.Run()`.
- Non-TTY / piped stdin or stdout → do **not** start Bubble Tea; print usage/help (same spirit as today’s bare `clocky`) and exit successfully unless an explicit error applies.
- `clocky help` remains explicit help text including a line that bare `clocky` opens the TUI when run in a terminal.

## Visual style

- Lip Gloss styles for borders, titles, active vs idle, selection highlight, errors.
- Box-drawing for panels; block characters for progress.
- **No emoji.**
- ASCII analog clock stays ASCII (existing `clockface`); session chrome may use box-drawing around it.
- Narrow terminals: stack panels vertically; shrink clock via existing width scaling; never panic.

## Error handling

| Case | Behavior |
|------|----------|
| Invalid form field | Block save; inline error string |
| Timer already running | Refuse start; point at Stop |
| Dead timer PID | Clear stale `timer.json`; show idle |
| Notify failure | Log/ignore at UI level; do not abort session |
| Resize | Relayout on `tea.WindowSizeMsg` |
| Quit mid-session | End pomodoro session; restore terminal; keep bg timer/stopwatch |

## Testing

- Unit-test TUI model `Update` handlers with Bubble Tea messages (tick, pause, key quits) using fake clocks where needed.
- Keep and extend domain tests (`presets`, `timer`, `stopwatch`, `duration`, `clockface`).
- Optional buffered Program smoke: hub view renders Active + Presets headers.
- CLI: non-TTY bare `clocky` does not hang; `go test ./...` green on Windows/macOS/Linux as today.

## Out of scope (this design)

- Mouse support
- User-configurable color themes / config file
- Multiple concurrent timers
- Session history or statistics
- Persisting / resuming an in-progress pomodoro after TUI quit
- Replacing or removing CLI subcommands
- Nerd Fonts / emoji icons

## Success criteria

1. `clocky` in a TTY opens the hub with live Active + Presets CRUD.
2. Timer and stopwatch can be started/stopped from the hub and keep running after quit; reopen shows correct live values.
3. Pomodoro session shows ASCII clock, supports pause/resume, notifications on phase end, return to hub.
4. Preset create/edit/delete/start works in TUI and stays compatible with CLI presets.
5. Visual style uses panels/colors/blocks without emoji; works on Windows, macOS, Linux terminals with ANSI.
6. `go test ./...` passes.

## Implementation touchpoints

- Primary: `internal/tui/**`, `internal/cli/cli.go` (empty-args → TUI)
- Secondary: thin adapters if timer/stopwatch APIs need hub-friendly read helpers; README help text
- Unchanged contracts preferred: `presets.json` schema, timer detach model, `clockface` API
)