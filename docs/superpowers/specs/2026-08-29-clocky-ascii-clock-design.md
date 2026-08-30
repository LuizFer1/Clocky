# Clocky — Larger ASCII Analog Clock Design

**Date:** 2026-08-29  
**Status:** Approved for planning  
**Scope:** Improve `internal/clockface` used by the foreground pomodoro display

## Goal

Make the pomodoro ASCII clock larger and more detailed, with distinct **minute** and **second** hands that reflect **remaining phase time**, plus cardinal tick labels at 12 / 3 / 6 / 9.

## Decisions (from brainstorming)

| Topic | Choice |
|-------|--------|
| Hand meaning | Remaining phase time (not wall clock, not single progress hand) |
| Default size | Medium ≈ **31×16** characters |
| Extra detail | Cardinal marks **12 / 3 / 6 / 9** + minute + second hands |
| Rendering approach | **A** — ASCII grid with two radial hands (no Unicode/braille) |

## Current behavior (baseline)

- `clockface.Render(remaining, total, width)` draws ≈21×11 ellipse rim (`o`), one progress hand from elapsed fraction of the phase, center `+`
- Pomodoro clears the screen each second and prints `Render` + `RenderCompact`
- Narrow terminals shrink via `faceSize(width)`

## Target behavior

### Face geometry

- **Default size:** width 31, height 16 when `width >= 31`
- **Scaling:** if terminal width &lt; 31, scale down proportionally (minimum ≈15×8); never panic
- Ellipse rim using ASCII (`o` / `.`)
- Cardinal labels at 12, 3, 6, 9 o’clock when space allows (literal `12` / `3` / `6` / `9`); otherwise a single mark character at those angles

### Hands (remaining time)

Angles: **0 at 12 o’clock, clockwise**.

```text
θ_s = 2π × (secs_remaining / 60)
θ_m = 2π × ((mins_remaining % 60) / 60)
```

- `secs_remaining` = remaining seconds modulo 60 (i.e. the SS part of MM:SS)
- `mins_remaining % 60` = minute hand within a 60-minute dial; phases longer than 60 minutes still show digital total remaining via `RenderCompact`
- **Minute hand:** length ≈ 70% of min(rx, ry); thicker stroke (`#` preferred, with `| / - \` by sector as fallback along the ray)
- **Second hand:** length ≈ 85% of min(rx, ry); thinner stroke (`*` preferred, with `| / - \` by sector)
- **Center:** `+`

`total` remains a `Render` parameter for API compatibility and tests; **hand angles depend only on `remaining`**.

### Cell overlap priority (highest wins)

1. Center  
2. Second hand  
3. Minute hand  
4. Cardinal mark / label  
5. Rim  
6. Space  

### Compact label

`RenderCompact` unchanged: `PHASE i/n MM:SS` below the face.

### Pomodoro integration

- Keep 1s redraw (`\033[H\033[2J` + `Render` + compact line)
- Pass terminal width into `Render` as today; default effective face grows with the new `faceSize`
- No change to timer/stopwatch (they do not use the analog face)

## API

```go
func Render(remaining, total time.Duration, width int) string
func RenderCompact(phase string, cycle, totalCycles int, remaining time.Duration) string
```

No new exported types required for MVP of this change.

## Testing

Update `internal/clockface/clockface_test.go`:

- Default-width render is at least ~31 columns and ~14 lines
- Output includes cardinal indication (e.g. contains `12` or agreed mark set)
- Different remainings (e.g. 25:00 vs 24:30 vs 0:05) produce different grids
- Narrow width still returns a non-empty sensible multi-line string
- Existing pomodoro tests that only assert “clockface output present” should still pass

## Out of scope

- Unicode / braille / box-drawing faces
- Hour hand
- Progress arc on the rim
- Fullscreen TUI frameworks
- Changing notification or timer background behavior

## Implementation touchpoints

- Primary: `internal/clockface/clockface.go`, `internal/clockface/clockface_test.go`
- Secondary (only if needed): default width in pomodoro hooks / CLI

## Success criteria

1. Pomodoro shows a clearly larger face than 21×11  
2. Second hand visibly moves each second  
3. Minute hand moves as remaining minutes change  
4. Marks at 12/3/6/9 are recognizable  
5. `go test ./...` passes
