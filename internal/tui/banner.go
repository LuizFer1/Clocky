package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Big block-letter CLOCKY using filled █ (5 rows × 5 cols per letter).
var clockyBannerLines = []string{
	" ███  █      ███   ███  █   █ █   █",
	"█     █     █   █ █     █  █   █ █ ",
	"█     █     █   █ █     ███     █  ",
	"█     █     █   █ █     █  █    █  ",
	" ███  █████  ███   ███  █   █   █  ",
}

// compactClocky for narrow terminals.
const compactClocky = "█▀▀ █   ▄▀▄ █▀▀ █▄▀ ▀▄▀"

func clockyBanner(width int) string {
	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color("51")).
		Bold(true)

	var art string
	if width > 0 && width < lipgloss.Width(clockyBannerLines[0])+4 {
		art = style.Render(compactClocky)
	} else {
		colored := make([]string, len(clockyBannerLines))
		for i, line := range clockyBannerLines {
			colored[i] = style.Render(line)
		}
		art = strings.Join(colored, "\n")
	}

	if width < 1 {
		return art
	}
	// Soft ░ fill behind / around the wordmark across the full width.
	pad := lipgloss.NewStyle().Foreground(lipgloss.Color("238")).Render
	rows := strings.Split(art, "\n")
	out := make([]string, 0, len(rows)+2)
	out = append(out, pad(strings.Repeat("░", width)))
	for _, row := range rows {
		centered := lipgloss.PlaceHorizontal(width, lipgloss.Center, row,
			lipgloss.WithWhitespaceChars("░"),
			lipgloss.WithWhitespaceForeground(lipgloss.Color("238")),
		)
		out = append(out, centered)
	}
	out = append(out, pad(strings.Repeat("░", width)))
	return strings.Join(out, "\n")
}
