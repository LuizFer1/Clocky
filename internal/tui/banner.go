package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Block-letter CLOCKY wordmark (half-block shading for a softer edge).
var clockyBannerLines = []string{
	"▄▀▀▀▄ ▄     ▄▀▀▀▄ ▄▀▀▀▄ ▄   ▄ ▄   ▄",
	"█     █     █   █ █     █ ▄▀   ▀▄▀ ",
	"█     █     █   █ █     █▀▄     █  ",
	"▀▄▄▄▀ ▀▄▄▄▄ ▀▄▄▄▀ ▀▄▄▄▀ ▀   ▀   ▀  ",
}

const bannerTagline = "focus · timers · stopwatch"

// compactClocky for narrow terminals.
const compactClocky = "█▀▀ █   ▄▀▄ █▀▀ █▄▀ ▀▄▀"

func clockyBanner(width int) string {
	style := lipgloss.NewStyle().Foreground(colAccent).Bold(true)
	tag := lipgloss.NewStyle().Foreground(colDim)

	var rows []string
	if width > 0 && width < lipgloss.Width(clockyBannerLines[0])+4 {
		rows = []string{style.Render(compactClocky)}
	} else {
		for _, line := range clockyBannerLines {
			rows = append(rows, style.Render(line))
		}
		rows = append(rows, tag.Render(bannerTagline))
	}

	if width < 1 {
		return strings.Join(rows, "\n")
	}
	out := make([]string, 0, len(rows)+2)
	out = append(out, "")
	for _, row := range rows {
		out = append(out, lipgloss.PlaceHorizontal(width, lipgloss.Center, row))
	}
	// Thin rule under the wordmark with a soft shaded end on each side.
	ends := tag.Render("░▒▓")
	mid := lipgloss.NewStyle().Foreground(colBorder).Render(strings.Repeat("─", max(0, width-6)))
	out = append(out, ends+mid+tag.Render("▓▒░"))
	return strings.Join(out, "\n")
}
