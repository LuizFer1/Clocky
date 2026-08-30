package clockface

import (
	"strings"
	"testing"
	"time"
)

func TestRenderFullRemainingHasRimAndLines(t *testing.T) {
	got := Render(25*time.Minute, 25*time.Minute, 21)
	lines := strings.Split(strings.TrimRight(got, "\n"), "\n")
	if len(lines) <= 5 {
		t.Fatalf("expected >5 lines, got %d:\n%s", len(lines), got)
	}
	if !strings.Contains(got, "o") && !strings.Contains(got, ".") {
		t.Fatalf("expected rim characters in:\n%s", got)
	}
	if !strings.Contains(got, "+") {
		t.Fatalf("expected center + in:\n%s", got)
	}
}

func TestRenderDifferentRemainingDiffer(t *testing.T) {
	a := Render(25*time.Minute, 25*time.Minute, 21)
	b := Render(12*time.Minute+30*time.Second, 25*time.Minute, 21)
	c := Render(0, 25*time.Minute, 21)
	if a == b {
		t.Fatal("full and half remaining produced identical faces")
	}
	if b == c {
		t.Fatal("half and zero remaining produced identical faces")
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
