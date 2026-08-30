package clockface

import (
	"fmt"
	"math"
	"strings"
	"time"
)

const (
	defaultWidth  = 35
	defaultHeight = 17
)

// Glyphs used by the face. Exported so the TUI can colorize by glyph.
const (
	RimRune    = '·'
	TickRune   = '•'
	CenterRune = '◉'
)

// MinuteHandRunes / SecondHandRunes are the glyph sets used for hands.
var (
	MinuteHandRunes = "━┃╱╲◆"
	SecondHandRunes = "─│/\\"
)

// Render returns a multi-line analog clock. Minute and second hands show
// remaining phase time (not wall clock). total is kept for API
// compatibility; hand angles depend only on remaining.
func Render(remaining, total time.Duration, width int) string {
	_ = total
	if remaining < 0 {
		remaining = 0
	}

	w, h := faceSize(width)
	grid := make([][]rune, h)
	for y := 0; y < h; y++ {
		grid[y] = runesRepeat(' ', w)
	}

	cx := float64(w-1) / 2
	cy := float64(h-1) / 2
	rx := cx
	ry := cy

	drawRim(grid, cx, cy, rx, ry)
	drawTicks(grid, cx, cy, rx, ry)
	drawCardinals(grid, cx, cy, rx, ry, w)

	totalSec := int64(remaining / time.Second)
	mins := totalSec / 60
	secs := totalSec % 60

	thetaM := 2 * math.Pi * float64(mins%60) / 60
	thetaS := 2 * math.Pi * float64(secs) / 60

	// Second hand is thin and long; minute hand heavy and shorter, both
	// inside the label ring so 12/3/6/9 stay readable.
	drawHand(grid, cx, cy, rx*0.80, ry*0.80, thetaS, SecondHandRunes, 0)
	drawHand(grid, cx, cy, rx*0.58, ry*0.58, thetaM, MinuteHandRunes, '◆')

	plot(grid, int(math.Round(cx)), int(math.Round(cy)), CenterRune)

	var b strings.Builder
	for y := 0; y < h; y++ {
		b.WriteString(string(grid[y]))
		b.WriteByte('\n')
	}
	return b.String()
}

// RenderCompact returns phase + cycle i/n + remaining MM:SS.
func RenderCompact(phase string, cycle, totalCycles int, remaining time.Duration) string {
	return fmt.Sprintf("%s %d/%d %s", phase, cycle, totalCycles, formatMMSS(remaining))
}

func faceSize(width int) (w, h int) {
	if width >= defaultWidth {
		return defaultWidth, defaultHeight
	}
	if width < 7 {
		width = 7
	}
	w = width
	if w%2 == 0 {
		w--
	}
	h = int(math.Round(float64(w) * float64(defaultHeight) / float64(defaultWidth)))
	if h < 5 {
		h = 5
	}
	if h%2 == 0 {
		h++
	}
	return w, h
}

// drawRim traces a single-thickness ellipse parametrically.
func drawRim(grid [][]rune, cx, cy, rx, ry float64) {
	steps := int(2*math.Pi*math.Max(rx, ry)) * 4
	if steps < 64 {
		steps = 64
	}
	for i := 0; i < steps; i++ {
		t := 2 * math.Pi * float64(i) / float64(steps)
		x := cx + rx*math.Sin(t)
		y := cy - ry*math.Cos(t)
		plot(grid, int(math.Round(x)), int(math.Round(y)), RimRune)
	}
}

// drawTicks places a stronger mark at each of the 12 hour positions.
func drawTicks(grid [][]rune, cx, cy, rx, ry float64) {
	for i := 0; i < 12; i++ {
		t := 2 * math.Pi * float64(i) / 12
		x := cx + rx*math.Sin(t)
		y := cy - ry*math.Cos(t)
		plot(grid, int(math.Round(x)), int(math.Round(y)), TickRune)
	}
}

func drawCardinals(grid [][]rune, cx, cy, rx, ry float64, faceW int) {
	if faceW < 21 {
		return
	}
	type card struct {
		label string
		theta float64
	}
	cards := []card{
		{"12", 0},
		{"3", math.Pi / 2},
		{"6", math.Pi},
		{"9", 3 * math.Pi / 2},
	}
	icx, icy := int(math.Round(cx)), int(math.Round(cy))
	for _, c := range cards {
		x := cx + rx*math.Sin(c.theta)
		y := cy - ry*math.Cos(c.theta)
		ix, iy := int(math.Round(x)), int(math.Round(y))
		// Snap to the true axes so 3/9 share a row and 12/6 a column.
		switch c.label {
		case "12", "6":
			ix = icx
		case "3", "9":
			iy = icy
		}
		start := ix - len(c.label)/2
		for i := -1; i <= len(c.label); i++ {
			plot(grid, start+i, iy, ' ')
		}
		for i, r := range c.label {
			plot(grid, start+i, iy, r)
		}
	}
}

// drawHand walks from the center outward choosing a glyph by slope.
// glyphs must be ordered horizontal, vertical, "/" diagonal, "\" diagonal.
func drawHand(grid [][]rune, cx, cy, lx, ly, theta float64, glyphs string, tip rune) {
	if lx < 1 && ly < 1 {
		return
	}
	set := []rune(glyphs)
	horiz, vert, slash, back := set[0], set[1], set[2], set[3]

	// Screen-space direction (cells are ~2:1, so ly is already half rx).
	dx := lx * math.Sin(theta)
	dy := -ly * math.Cos(theta)
	// Slope in cell units; visually compare after un-squashing rows.
	ax, ay := math.Abs(dx), math.Abs(dy)*2
	var ch rune
	switch {
	case ay < ax*0.45:
		ch = horiz
	case ax < ay*0.45:
		ch = vert
	case (dx > 0) == (dy < 0):
		ch = slash
	default:
		ch = back
	}

	// One glyph per column for horizontal hands, one per row otherwise,
	// so diagonals never double up.
	steps := int(math.Round(math.Abs(dy)))
	if ch == horiz {
		steps = int(math.Round(math.Abs(dx)))
	}
	if steps < 1 {
		steps = 1
	}
	var lastX, lastY = -1, -1
	for i := 1; i <= steps; i++ {
		t := float64(i) / float64(steps)
		x := int(math.Round(cx + dx*t))
		y := int(math.Round(cy + dy*t))
		if x == lastX && y == lastY {
			continue
		}
		lastX, lastY = x, y
		plot(grid, x, y, ch)
	}
	if tip != 0 && lastX >= 0 {
		plot(grid, lastX, lastY, tip)
	}
}

func plot(grid [][]rune, x, y int, ch rune) {
	if y < 0 || y >= len(grid) || x < 0 || x >= len(grid[0]) {
		return
	}
	grid[y][x] = ch
}

func formatMMSS(d time.Duration) string {
	if d < 0 {
		d = -d
	}
	totalSec := int64(d / time.Second)
	m := totalSec / 60
	s := totalSec % 60
	return fmt.Sprintf("%02d:%02d", m, s)
}

func runesRepeat(c rune, n int) []rune {
	b := make([]rune, n)
	for i := range b {
		b[i] = c
	}
	return b
}
