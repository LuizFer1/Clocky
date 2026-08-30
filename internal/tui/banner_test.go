package tui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestClockyBannerContainsBlocks(t *testing.T) {
	b := clockyBanner(80)
	if !strings.Contains(b, "█") {
		t.Fatal("expected filled block characters")
	}
	if !strings.Contains(b, "░") {
		t.Fatal("expected shade fill around banner")
	}
	if lipgloss.Width(b) > 80 {
		t.Fatalf("width=%d", lipgloss.Width(b))
	}
}

func TestClockyBannerNarrow(t *testing.T) {
	b := clockyBanner(20)
	if !strings.Contains(b, "█") {
		t.Fatal("expected blocks in compact banner")
	}
}
