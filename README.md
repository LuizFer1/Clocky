# Clocky

A simple cross-platform terminal time manager for **Windows**, **macOS**, and **Linux**.

Clocky gives you a Pomodoro timer with an ASCII analog clock, a background countdown timer with system notifications, and a toggle stopwatch — from an interactive Terminal UI or the classic CLI.

```text
clocky                 # Terminal UI hub (interactive terminal)
clocky pomodoro
clocky timer :25:
clocky stopwatch
```

## Features

- **Terminal UI** — hub with live timer/stopwatch, preset CRUD, and pomodoro sessions (ASCII clock + pause)
- **Pomodoro** — focus / short break / long break cycles with a live ASCII analog clock face
- **Timer** — runs in the background and notifies you when time is up (beep + banner + native OS notification)
- **Stopwatch** — start on first call, stop and print elapsed time on the second
- **Named presets** — save custom pomodoros and timers in the TUI or via `clocky add`
- **Flexible duration syntax** — express time as `H:M:S` with optional empty fields

## Install

### One-line install

**Windows (PowerShell):**

```powershell
irm https://raw.githubusercontent.com/LuizFer1/Clocky/main/scripts/install.ps1 | iex
```

**macOS / Linux:**

```bash
curl -fsSL https://raw.githubusercontent.com/LuizFer1/Clocky/main/scripts/install.sh | bash
```

Defaults: Windows `%LOCALAPPDATA%\Clocky`, Unix `~/.local/bin`. Override with `CLOCKY_INSTALL_DIR`, pin a release with `CLOCKY_VERSION=v1.2.0`, or point at a fork with `CLOCKY_GITHUB_REPO=owner/name`.

### Update

```bash
clocky update
clocky update --yes
clocky update --check
```

### From source

Requires [Go](https://go.dev/) 1.22+.

```bash
git clone https://github.com/LuizFer1/Clocky.git
cd Clocky
go build -o clocky ./cmd/clocky
```

Move the binary somewhere on your `PATH`, for example:

```bash
# Linux / macOS
sudo mv clocky /usr/local/bin/

# Windows (PowerShell, example)
Move-Item .\clocky.exe $env:USERPROFILE\bin\clocky.exe
```

## Quick start

```bash
# Interactive Terminal UI (hub: active timers + presets)
clocky

# Classic 25/5 pomodoro (ASCII clock in the terminal)
clocky pomodoro

# Custom pomodoro
clocky pomodoro --focus 40 --break 10 --long 15 --cycles 4 --auto

# Background timer (25 minutes)
clocky timer :25:

# Stopwatch toggle
clocky stopwatch   # start
clocky stopwatch   # stop → prints elapsed HH:MM:SS

# What's running?
clocky status
```

## Commands

| Command | Description |
|---------|-------------|
| `clocky` | Open the Terminal UI (TTY); prints help when not interactive |
| `clocky pomodoro [name] [flags]` | Start a pomodoro (defaults or named preset) |
| `clocky timer <duration\|name>` | Start a background timer |
| `clocky timer --stop` | Cancel the active timer |
| `clocky stopwatch` | Toggle stopwatch start / stop |
| `clocky add pomodoro ...` | Save a named pomodoro preset |
| `clocky add timer ...` | Save a named timer preset |
| `clocky list [pomodoro\|timer]` | List saved presets |
| `clocky remove <type> <name>` | Delete a preset |
| `clocky status` | Show active timer / stopwatch |
| `clocky version` | Print embedded version |
| `clocky update [--yes] [--check]` | Upgrade from GitHub Releases |
| `clocky help` | Show usage |

### Pomodoro flags

| Flag | Default | Description |
|------|---------|-------------|
| `--focus` | `25` | Focus duration (minutes) |
| `--break` | `5` | Short break (minutes) |
| `--long` | `15` | Long break (minutes) |
| `--cycles` | `4` | Focus sessions before a long break |
| `--auto` | off | Advance phases without pressing Enter |

Between phases, Clocky beeps, prints an ASCII banner, and sends a **native OS notification**. Press Enter to continue (or use `--auto`).

## Named presets

Save configurations once, reuse them by name:

```bash
clocky add pomodoro --name Focus --focus 40 --break 10 --long 15 --cycles 4 --auto
clocky add timer --name Break ::300

clocky pomodoro Focus
clocky timer Break
```

```bash
clocky list
clocky remove pomodoro Focus
clocky remove timer Break
```

## Duration format

Durations use `H:M:S`. Empty fields are allowed; values overflow into larger units.

| Input | Meaning |
|-------|---------|
| `1:30:00` | 1 hour 30 minutes |
| `24::` | 24 hours |
| `:25:` | 25 minutes |
| `::50` | 50 seconds |
| `:1440:` | 1440 minutes → **24 hours** |
| `::1440` | 1440 seconds → **24 minutes** |
| `90` | 90 seconds |

## How it works

- Single Go binary — no daemon required
- State lives in `~/.clocky/` (`presets.json`, `timer.json`, `stopwatch.json`)
- Pomodoro runs in the **foreground** with an in-place ASCII clock
- Timer spawns a **background** process and notifies on completion

## Notifications

When a pomodoro phase ends or a timer finishes, Clocky:

1. Beeps in the terminal
2. Prints an ASCII banner
3. Shows a native notification (Windows toast, macOS Notification Center, or `notify-send` on Linux)

## Development

Use a separate binary name so local builds do not overwrite an installed release `clocky`:

```bash
go test ./...

# Windows (PowerShell)
.\scripts\build-dev.ps1
.\clockyDEV.exe pomodoro --focus 1 --cycles 1 --auto

# macOS / Linux
./scripts/build-dev.sh
./clockyDEV pomodoro --focus 1 --cycles 1 --auto
```

Or manually: `go build -o clockyDEV.exe ./cmd/clocky` (Windows) / `go build -o clockyDEV ./cmd/clocky` (Unix).

Design notes: [`docs/superpowers/specs/2026-08-29-clocky-design.md`](docs/superpowers/specs/2026-08-29-clocky-design.md)

## License

MIT (planned)
