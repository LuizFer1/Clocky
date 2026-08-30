# Clocky — Design Spec

**Date:** 2026-08-29  
**Status:** Approved for planning  
**Platform:** Windows, macOS, Linux (terminal CLI)

## Goal

Clocky is a simple cross-platform terminal time-management tool with:

- Pomodoro (foreground, ASCII analog clock)
- Timer (background, notifies on completion)
- Stopwatch / cronômetro (toggle start/stop)
- Named presets via `clocky add`

## Approach

**CLI modular with file-based state (Approach A).**

Single Go binary, subcommands, persistent state under `~/.clocky/`. No always-on daemon. No fullscreen TUI framework required for MVP (ANSI redraw only for pomodoro).

## Architecture

```
clocky
├── cmd/clocky/           # main entrypoint, flag/subcommand routing
├── internal/
│   ├── pomodoro/         # cycle logic, phase transitions
│   ├── timer/            # background countdown, stop, completion
│   ├── stopwatch/        # cronômetro start/stop toggle
│   ├── presets/          # add / list / remove named pomodoros & timers
│   ├── duration/         # flexible H:M:S parsing and normalization
│   ├── clockface/        # ASCII analog clock face renderer
│   ├── state/            # ~/.clocky/ JSON + PID helpers
│   └── notify/           # beep + ASCII banner + native OS notification
└── go.mod
```

### State layout (`~/.clocky/`)

| Path | Purpose |
|------|---------|
| `presets.json` | Named pomodoro and timer presets |
| `stopwatch.json` | Active stopwatch (`started_at`) or absent/empty when stopped |
| `timer.json` | Active background timer (deadline, PID, name/label) |

## Commands

### `clocky pomodoro [name] [flags]`

- **Foreground.** Renders ASCII analog clock + remaining `MM:SS` + phase + cycle.
- If `name` is given, load preset from `presets.json` (flags may override preset fields).
- If no name, use defaults or inline flags.

**Defaults:** `--focus 25` (minutes), `--break 5`, `--long 15`, `--cycles 4`

**Flags:**

| Flag | Meaning |
|------|---------|
| `--focus <min>` | Focus duration in minutes |
| `--break <min>` | Short break duration in minutes |
| `--long <min>` | Long break duration in minutes |
| `--cycles <n>` | Focus sessions before a long break |
| `--auto` | Advance phases without waiting for Enter |

**Flow:** focus → notify → (Enter or `--auto`) → short break → … → after N focus sessions → long break → repeat or exit after configured cycles as implemented (MVP: one full set of `cycles` focus sessions, then exit after final long break unless user restarts).

**Cancel:** `Ctrl+C` stops cleanly and prints a short message.

### `clocky timer <duration|name> [--stop]`

- **Background.** Detaches a worker process; parent returns immediately.
- Argument is either a **duration** (flexible parse) or a **preset name**.
- On completion: beep + ASCII banner (if TTY available) + native OS notification.
- `clocky timer --stop` cancels the active timer (kill PID, clear `timer.json`).
- If a timer is already running, refuse to start another and point user to `status` / `--stop`.

**Examples:**

```text
clocky timer ::50
clocky timer :25:
clocky timer 24::
clocky timer Break
clocky timer --stop
```

### `clocky cronometro`

- **Toggle.** First call writes `started_at` to `stopwatch.json` and prints that it started.
- Second call computes elapsed, prints `HH:MM:SS`, clears state.
- No live ticking display.

### `clocky add pomodoro` / `clocky add timer`

Persist named presets.

```text
clocky add pomodoro --name Focus --focus 40 --break 10 --long 15 --cycles 4 --auto
clocky add timer --name Break ::300
```

- `--name` is required and unique within its type (case-sensitive as stored; lookup exact match).
- Pomodoro fields mirror pomodoro flags (`--focus`, `--break`, `--long`, `--cycles`, `--auto`).
- Timer requires a duration argument after flags (flexible format).
- Overwriting an existing name: replace and print a short confirmation (MVP: allow overwrite with message).

Then:

```text
clocky pomodoro Focus
clocky timer Break
```

### `clocky list` / `clocky remove`

```text
clocky list                 # all presets
clocky list pomodoro
clocky list timer
clocky remove pomodoro Focus
clocky remove timer Break
```

### `clocky status`

Shows:

- Active stopwatch elapsed (if any)
- Active timer remaining + label (if any)

### `clocky help` / bare `clocky`

Print usage in Portuguese.

## Duration parsing (`internal/duration`)

Canonical storage: total seconds (`time.Duration`).

**Accepted forms:**

| Input | Meaning | Example result |
|-------|---------|----------------|
| `H:M:S` | hours, minutes, seconds | `1:30:00` → 1h30m |
| `H:M:` | hours, minutes, 0 seconds | `1:30:` → 1h30m |
| `H::S` | hours, 0 minutes, seconds | `1::30` → 1h30s |
| `:M:S` | 0 hours, minutes, seconds | `:5:30` → 5m30s |
| `H::` | hours only | `24::` → 24h |
| `:M:` | minutes only | `:25:` → 25m |
| `::S` | seconds only | `::50` → 50s |
| bare integer | seconds | `90` → 90s |

**Overflow normalization (required):**

- `:1440:` → 1440 minutes → **24 hours**
- `::1440` → 1440 seconds → **24 minutes**
- `:90:00` → 90 minutes → **1 hour 30 minutes** (when displaying)

Display helper always prints normalized `HH:MM:SS` (hours may exceed 24 if needed, no day field in MVP).

Invalid input → non-zero exit and Portuguese usage hint.

**Note:** Flag spelling is `--break` (not `--breack`).

## Notifications (`internal/notify`)

MVP stack (all three):

1. **Terminal beep** (BEL / platform beep)
2. **ASCII banner** in the terminal (phase change / timer done)
3. **Native OS notification**
   - Windows: toast via suitable Go library or PowerShell fallback
   - macOS: `osascript` notification
   - Linux: `notify-send` when available; degrade gracefully if missing

Pomodoro transitions always notify. Timer completion always notifies (native is critical because process may be backgrounded).

Between pomodoro phases without `--auto`, wait for Enter after notification.

## ASCII analog clock (`internal/clockface`)

- Face ~21×11 characters with minute (and second if readable) hands recomputed each tick
- Below face: phase label (`FOCO` / `PAUSA` / `PAUSA LONGA`), cycle `i/n`, remaining `MM:SS`
- Redraw in-place via ANSI cursor home/clear-line (no infinite scroll)
- Narrow terminal fallback: phase + digits only, skip face

## Error handling

Messages in Portuguese. Non-zero exit on invalid usage.

Examples:

- Bad duration → `uso: clocky timer <H:M:S|nome>`
- Timer already running → suggest `clocky status` and `clocky timer --stop`
- Unknown preset name → list hint via `clocky list`
- Missing `--name` on add → usage error

## Testing

**Unit tests:**

- Duration parse matrix (including `:1440:`, `::1440`, `24::`, `:25:`, `::50`)
- Stopwatch start/stop state transitions
- Preset add / get / overwrite / remove
- Pomodoro phase sequence and cycle boundaries
- Clock hand angle → character cell mapping (table-driven)

**Integration (light):**

- CLI: add preset → run by name (pomodoro with 1s focus in test helper / injectable clock)
- CLI: timer 1s background → status → completion path (notify mocked)
- CLI: cronometro twice

Do not require a real desktop notification server in CI; mock `notify` interface.

## Out of scope (MVP)

- Always-on daemon
- Fullscreen TUI framework (Bubble Tea etc.)
- TTS / `--speak`
- Multiple concurrent timers
- Cloud sync / accounts
- GUI

## Success criteria

1. `go build` produces a single runnable binary on Windows, macOS, and Linux.
2. Documented commands behave as specified.
3. Flexible duration parsing matches the conversion table.
4. Named presets round-trip via `add` → invoke by name.
5. Pomodoro shows ASCII analog countdown; timer notifies on completion with beep + banner + native notify.
6. Unit tests cover duration, presets, and pomodoro cycles.
