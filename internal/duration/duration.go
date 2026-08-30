package duration

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Parse accepts flexible H:M:S forms (empty fields allowed) or a bare integer of seconds.
// Overflow is allowed (e.g. :1440: is 24h). Durations <= 0 are rejected.
func Parse(s string) (time.Duration, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, fmt.Errorf("empty duration")
	}
	if !strings.Contains(s, ":") {
		sec, err := strconv.Atoi(s)
		if err != nil || sec <= 0 {
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
	total := time.Duration(h)*time.Hour + time.Duration(m)*time.Minute + time.Duration(sec)*time.Second
	if total <= 0 {
		return 0, fmt.Errorf("duration must be > 0")
	}
	return total, nil
}

// Format renders d as HH:MM:SS. Hours may exceed 24.
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
