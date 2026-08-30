package clockface

import (
	"fmt"
	"math"
	"strings"
	"time"
)

const (
	defaultWidth  = 31
	defaultHeight = 16
)

// Render returns a multi-line ASCII analog clock. Minute and second hands
// show remaining phase time (not wall clock). total is kept for API
// compatibility; hand angles depend only on remaining.
func Render(remaining, total time.Duration, width int) string {
	_ = total
	if remaining < 0 {
		remaining = 0
	}

	w, h := faceSize(width)
	grid := make([][]byte, h)
	for y := 0; y < h; y++ {
		grid[y] = bytesRepeat(byte(' '), w)
	}

	cx := float64(w-1) / 2
	cy := float64(h-1) / 2
	rx := cx
	ry := cy
	rMin := math.Min(rx, ry)

	drawRim(grid, cx, cy, rx, ry)
	drawCardinals(grid, cx, cy, rx, ry, w)

	totalSec := int64(remaining / time.Second)
	mins := totalSec / 60
	secs := totalSec % 60

	thetaM := 2 * math.Pi * float64(mins%60) / 60
	thetaS := 2 * math.Pi * float64(secs) / 60

	// Keep hands inside the label ring so 12/3/6/9 stay readable.
	drawHand(grid, cx, cy, rMin*0.62, thetaM, '#')
	drawHand(grid, cx, cy, rMin*0.78, thetaS, '*')

	plot(grid, int(math.Round(cx)), int(math.Round(cy)), '+')

	var b strings.Builder
	for y := 0; y < h; y++ {
		b.Write(grid[y])
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
	h = int(math.Round(float64(w) * float64(defaultHeight) / float64(defaultWidth)))
	if h < 5 {
		h = 5
	}
	if h%2 == 0 {
		h++
	}
	return w, h
}

func drawRim(grid [][]byte, cx, cy, rx, ry float64) {
	h := len(grid)
	w := len(grid[0])
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			dx := (float64(x) - cx) / rx
			dy := (float64(y) - cy) / ry
			d := dx*dx + dy*dy
			if d >= 0.86 && d <= 1.14 {
				grid[y][x] = 'o'
			}
		}
	}
}

func drawCardinals(grid [][]byte, cx, cy, rx, ry float64, faceW int) {
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
	// Place labels on the outer ring (beyond hand tips).
	fr := 0.98
	useLabels := faceW >= 21
	icx, icy := int(math.Round(cx)), int(math.Round(cy))
	for _, c := range cards {
		x := cx + rx*fr*math.Sin(c.theta)
		y := cy - ry*fr*math.Cos(c.theta)
		ix, iy := int(math.Round(x)), int(math.Round(y))
		// Snap cardinals to the true axes so 3/9 share a row and 12/6 a column.
		switch c.label {
		case "12", "6":
			ix = icx
		case "3", "9":
			iy = icy
		}
		if !useLabels {
			plot(grid, ix, iy, 'o')
			continue
		}
		label := c.label
		start := ix - len(label)/2
		// Clear a small pad so digits don't glue into the rim.
		for i := -1; i <= len(label); i++ {
			plot(grid, start+i, iy, ' ')
		}
		for i := 0; i < len(label); i++ {
			plot(grid, start+i, iy, label[i])
		}
	}
}

func drawHand(grid [][]byte, cx, cy, length, theta float64, ch byte) {
	if length < 1 {
		return
	}
	steps := int(length*2) + 2
	for i := 1; i <= steps; i++ {
		t := float64(i) / float64(steps)
		x := cx + length*t*math.Sin(theta)
		y := cy - length*t*math.Cos(theta)
		plot(grid, int(math.Round(x)), int(math.Round(y)), ch)
	}
}

func plot(grid [][]byte, x, y int, ch byte) {
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

func bytesRepeat(c byte, n int) []byte {
	b := make([]byte, n)
	for i := range b {
		b[i] = c
	}
	return b
}
