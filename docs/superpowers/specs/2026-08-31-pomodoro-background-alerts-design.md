# Clocky — Pomodoro Background + Alerts Fix

**Date:** 2026-08-31  
**Status:** Approved for planning  
**Platform:** Windows, macOS, Linux (interactive TUI)  
**Related:** `2026-08-30-clocky-terminal-ui-design.md`

## Goal

Fix the broken end-of-phase experience in the TUI pomodoro session, and keep a started pomodoro running when the user leaves the ASCII session view for the hub.

Concrete outcomes:

1. **Visual bug fixed** — phase-end no longer corrupts the terminal (no stdout banner during Bubble Tea).
2. **Audible + visual alerts** — same hub pattern as finished timers: orange notice, repeating native beep until Esc, plus desktop toast.
3. **Background in-app** — Esc from the session minimizes to the hub; the cycle continues; Active shows phase/remaining; user can reopen the clock or stop the session from the hub.

## Decisions (from brainstorming)

| Topic | Choice |
|-------|--------|
| Survive leaving session view (Esc → hub) | Yes |
| Survive quitting the whole Clocky process | No (in-memory only; same as today for pomodoro) |
| Alert UX | Match timer hub: notice + repeating `notify.Alert()` until Esc; also desktop toast |
| Next phase when ending on hub | Respect preset `Auto` (auto-advance vs waiting) |
| Hub controls | Layout A: Active line + action buttons **Open Pomodoro** / **Stop Pomodoro** |
| Esc in session | Minimize to hub (does **not** cancel) |
| Approach | In-process live session owned by `appModel` (not disk worker) |

## Problem diagnosis

Today `sessionModel.finishPhase` calls `notify.All`, whose `Default.Banner` writes `*** title ***` to stdout. Inside Bubble Tea that paints outside the model `View()`, scrambling the ASCII face (observed as yellow/cyan digit garbage at `00:00`). Beep via `\a` is weak/unreliable in this context; the hub timer path already uses `notify.Alert()` (MessageBeep on Windows) and a persistent notice panel — pomodoro does not.

Pomodoro progress lives only while `page == pageSession`. Esc emits `sessionAbortMsg`, which drops the session and returns to hub, so there is no background continuation.

## Approach

**A — Live in-app session (chosen).** Keep one `sessionModel` (or equivalent runtime state) on `appModel` for the lifetime of the TUI run. Page switches do not destroy it. Hub ticks advance the session when the user is not on the session page. No `pomodoro.json`, no detached worker.

Rejected for this change:

- **B — Disk persistence** — survives process exit; out of scope.
- **C — Reuse timer worker per phase** — pause / multi-phase / Auto fit poorly.

## Architecture

```text
appModel
├── page (hub | session | …)
├── hub
└── session (always retained while a run is active)
      ├── phases / index / remaining / paused / waitingEnter
      └── cfg (incl. Auto)

Hub tick (1s)
  → if session active && page != session: advance session clock
  → on phase end: set notice + alerting + desktop toast (no stdout banner)
  → refresh Active snapshot to include pomodoro line

Esc on session page
  → page = hub (minimize); session keeps running

Open Pomodoro (hub action)
  → page = session (same sessionModel)

Stop Pomodoro (hub action)
  → clear session; stop alert; Active shows idle
```

### Notification rules (TUI)

| Channel | Behavior |
|---------|----------|
| Stdout banner (`notify.Default.Banner`) | **Do not use** during TUI |
| Terminal bell (`\a`) | Optional / unused for hub-style alerts |
| `notify.Alert()` / `StopAlert()` | Repeating while `alerting` (same as timer) |
| Hub `notice` panel | Phase-complete message until Esc dismisses |
| `notify.Desktop` / `Notifier.Desktop` | One toast per phase end |

CLI `clocky pomodoro` may keep using `notify.All` (banner is appropriate outside Bubble Tea).

## Components

### `appModel` / session ownership

- Starting a pomodoro creates/resets `session` and sets `pageSession`.
- `sessionDoneMsg` (natural end of last phase): clear active session, hub notice already shown, stay/return hub.
- Replace abort-on-Esc with **minimize**: no longer treat Esc as “destroy session”.
- Starting a new pomodoro while one is active: refuse with hub error (mirror “timer already running”).

### Hub Active + actions

- Active line when running, e.g. `● Pomodoro  FOCUS 1/4  12:34 left` (or `waiting` / `PAUSED` when applicable).
- Action bar gains **Open Pomodoro** and **Stop Pomodoro** when a session is active (Layout A). Existing New Pomodoro / timer / stopwatch actions remain.
- Footer: document Esc stops alarm; Open/Stop bindings as implemented.

### Session view

- Space still pauses/resumes **only while on the session page**. While minimized on the hub, remaining time continues (pause does not apply from hub).
- Phase end on the session page: fire the same alert pipeline (notice is owned by hub state so it is visible after minimize; if still on session, either show an in-session status line **and** set hub notice for when they return, or route alert through app so hub fields update). Prefer: alert state lives on hub/`appModel` so one dismiss path (Esc on hub) works.
- `n` / Enter advance behavior unchanged relative to Auto / waitingEnter.
- Footer: Esc/b → hub (minimize), not “end session”. Stopping is explicit via hub **Stop Pomodoro** (and optionally a dedicated key later — not required if button exists).

### Visual integrity

- Never call Banner-to-stdout from TUI phase completion.
- Rendering `remaining == 0` is fine if colorFace only runs inside `View()`; avoid mixing foreign stdout with the frame.

## Data flow

```text
Start pomodoro
  → new sessionModel; page = session

Each second (session page)
  → session.Update(tick) as today

Each second (hub page, session active)
  → app applies the same remaining decrement / phase-end logic
  → hub Active shows live pomodoro fields
  → phase end → notice + alerting + Desktop(title, body)

Esc on session
  → page = hub; session untouched

Open Pomodoro
  → page = session

Stop Pomodoro
  → session cleared; alerting false; StopAlert()

q / Ctrl+C
  → quit TUI; pomodoro ends with process (not persisted)
```

## Error handling

| Case | Behavior |
|------|----------|
| Second pomodoro start while active | Refuse; status/error on hub |
| Notify / Desktop failure | Ignore at UI level; do not abort session |
| Alert while on session page | Still set hub notice/alerting; beep runs; user can Esc after returning to hub, or dismiss if Esc on session clears alerting without minimizing when a notice is active (prefer: Esc dismisses alarm first if alerting, else minimizes — same priority as hub today) |
| Auto=false phase end on hub | `waitingEnter`; Active shows waiting; Open or Enter advances |

## Testing

- Model tests: minimize keeps remaining; hub tick crosses zero → notice + waiting/auto advance; Stop clears session; refuse double start.
- Alert path: TUI phase end must not invoke stdout Banner (unit test with fake Notifier or assert helper used by session).
- Existing timer notice/alert tests remain green.
- `go test ./...` on Windows.

## Out of scope

- Persisting / resuming pomodoro after quitting Clocky
- Detached pomodoro worker / `pomodoro.json`
- Changing CLI pomodoro banner behavior
- Mouse support, themes, session history

## PR plan (high level)

1. Introduce app-owned session lifecycle (minimize / open / stop) and hub Active + action buttons.
2. Wire background tick + Auto/waiting on hub; phase-end alerts via notice + `Alert` + Desktop (no Banner).
3. Tests for lifecycle, alerts, and regression on timer hub alarm.
