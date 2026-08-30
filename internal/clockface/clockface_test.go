package clockface

import (
	"strings"
	"testing"
	"time"
)

func TestRenderDefaultSizeIsLarger(t *testing.T) {
	got := Render(25*time.Minute, 25*time.Minute, 80)
	lines := strings.Split(strings.TrimRight(got, "\n"), "\n")
	if len(lines) < 14 {
		t.Fatalf("expected >=14 lines, got %d:\n%s", len(lines), got)
	}
	if len(lines[0]) < 31 {
		t.Fatalf("expected width >=31, got %d:\n%s", len(lines[0]), got)
	}
	if !strings.Contains(got, "o") && !strings.Contains(got, ".") {
		t.Fatalf("expected rim characters in:\n%s", got)
	}
	if !strings.Contains(got, "+") {
		t.Fatalf("expected center + in:\n%s", got)
	}
}

func TestRenderHasCardinalMarks(t *testing.T) {
	// Hands away from 12/3/6/9: ~20 min and ~10 sec.
	got := Render(20*time.Minute+10*time.Second, 25*time.Minute, 80)
	for _, mark := range []string{"12", "3", "6", "9"} {
		if !strings.Contains(got, mark) {
			t.Fatalf("missing cardinal %q in:\n%s", mark, got)
		}
	}
}

func TestRenderDifferentRemainingDiffer(t *testing.T) {
	a := Render(25*time.Minute, 25*time.Minute, 80)
	b := Render(24*time.Minute+30*time.Second, 25*time.Minute, 80)
	c := Render(5*time.Second, 25*time.Minute, 80)
	if a == b {
		t.Fatal("25:00 and 24:30 produced identical faces")
	}
	if b == c {
		t.Fatal("24:30 and 0:05 produced identical faces")
	}
	if a == c {
		t.Fatal("25:00 and 0:05 produced identical faces")
	}
}

func TestRenderSecondHandMovesEachSecond(t *testing.T) {
	a := Render(1*time.Minute+10*time.Second, time.Minute*2, 80)
	b := Render(1*time.Minute+9*time.Second, time.Minute*2, 80)
	if a == b {
		t.Fatal("second hand did not move between :10 and :09")
	}
}

func TestRenderNarrowStillSensible(t *testing.T) {
	got := Render(time.Minute, 2*time.Minute, 9)
	if strings.TrimSpace(got) == "" {
		t.Fatal("narrow render empty")
	}
	lines := strings.Split(strings.TrimRight(got, "\n"), "\n")
	if len(lines) < 3 {
		t.Fatalf("narrow face too small (%d lines):\n%s", len(lines), got)
	}
}

func TestRenderCompactContainsPhaseAndTime(t *testing.T) {
	got := RenderCompact("FOCUS", 2, 4, 5*time.Minute+7*time.Second)
	if !strings.Contains(got, "FOCUS") {
		t.Fatalf("missing phase: %q", got)
	}
	if !strings.Contains(got, "2/4") {
		t.Fatalf("missing cycle: %q", got)
	}
	if !strings.Contains(got, "05:07") {
		t.Fatalf("missing MM:SS: %q", got)
	}
}

