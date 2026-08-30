package clockface

import (
	"fmt"
	"math"
	"strings"
	"time"
)

// Render returns a multi-line ASCII analog clock. Hands move with
// progress = 1 - remaining/total (elapsed fraction of the phase).
func Render(remaining, total time.Duration, width int) string {
	if total <= 0 {
		total = time.Second
	}
	if remaining < 0 {
		remaining = 0
	}
	if remaining > total {
		remaining = total
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

	// Ellipse rim.
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			dx := (float64(x) - cx) / rx
			dy := (float64(y) - cy) / ry
			d := dx*dx + dy*dy
			if d >= 0.82 && d <= 1.18 {
				grid[y][x] = 'o'
			}
		}
	}

	progress := 1 - float64(remaining)/float64(total)
	theta := 2 * math.Pi * progress // 0 at 12 o'clock, clockwise

	handLen := math.Min(rx, ry) * 0.75
	steps := int(handLen) + 2
	for i := 1; i <= steps; i++ {
		t := float64(i) / float64(steps)
		x := cx + handLen*t*math.Sin(theta)
		y := cy - handLen*t*math.Cos(theta)
		ix, iy := int(math.Round(x)), int(math.Round(y))
		if iy >= 0 && iy < h && ix >= 0 && ix < w && grid[iy][ix] != 'o' {
			grid[iy][ix] = handChar(theta)
		}
	}

	icx, icy := int(math.Round(cx)), int(math.Round(cy))
	if icy >= 0 && icy < h && icx >= 0 && icx < w {
		grid[icy][icx] = '+'
	}

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
	if width >= 21 {
		return 21, 11
	}
	if width < 7 {
		width = 7
	}
	w = width
	h = w / 2
	if h < 5 {
		h = 5
	}
	if h%2 == 0 {
		h++
	}
	return w, h
}

func handChar(theta float64) byte {
	// Normalize to [0, 2π).
	for theta < 0 {
		theta += 2 * math.Pi
	}
	for theta >= 2*math.Pi {
		theta -= 2 * math.Pi
	}
	sector := int(math.Floor((theta+math.Pi/8)/(math.Pi/4))) % 8
	switch sector {
	case 0, 4:
		return '|'
	case 1, 5:
		return '/'
	case 2, 6:
		return '-'
	default:
		return '\\'
	}
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
