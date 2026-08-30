package tui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestCenterInFitsCanvas(t *testing.T) {
	out := centerIn(40, 10, "hi")
	if lipgloss.Width(out) > 40 {
		t.Fatalf("width=%d", lipgloss.Width(out))
	}
	if lipgloss.Height(out) != 10 {
		t.Fatalf("height=%d", lipgloss.Height(out))
	}
	if !strings.Contains(out, "hi") {
		t.Fatalf("missing content: %q", out)
	}
}

func TestFillFrameUsesFullHeight(t *testing.T) {
	out := fillFrame(60, 20, "Clocky", "BODY", "help")
	if lipgloss.Height(out) != 20 {
		t.Fatalf("height=%d want 20\n%s", lipgloss.Height(out), out)
	}
	if !strings.Contains(out, "Clocky") || !strings.Contains(out, "BODY") || !strings.Contains(out, "help") {
		t.Fatalf("missing parts:\n%s", out)
	}
}
