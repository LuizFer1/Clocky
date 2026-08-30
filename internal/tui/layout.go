package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// centerIn places content in the middle of a w×h canvas.
func centerIn(w, h int, content string) string {
	if w < 1 {
		w = 1
	}
	if h < 1 {
		h = 1
	}
	return lipgloss.Place(w, h, lipgloss.Center, lipgloss.Center, content,
		lipgloss.WithWhitespaceForeground(lipgloss.Color("236")),
	)
}

// fillFrame paints a full-terminal frame: optional header, body (expands), footer.
func fillFrame(w, h int, header, body, footer string) string {
	if w < 1 {
		w = 80
	}
	if h < 1 {
		h = 24
	}

	headerStyle := lipgloss.NewStyle().
		Width(w).
		Padding(0, 2).
		Bold(true).
		Foreground(lipgloss.Color("231")).
		Background(lipgloss.Color("57"))

	footerStyle := lipgloss.NewStyle().
		Width(w).
		Padding(0, 2).
		Foreground(lipgloss.Color("250")).
		Background(lipgloss.Color("236"))

	bodyStyle := lipgloss.NewStyle().
		Width(w).
		Foreground(lipgloss.Color("252"))

	head := headerStyle.Render(header)
	foot := footerStyle.Render(footer)

	used := lipgloss.Height(head) + lipgloss.Height(foot)
	bodyH := h - used
	if bodyH < 1 {
		bodyH = 1
	}

	inner := bodyStyle.Height(bodyH).Render(
		lipgloss.Place(w, bodyH, lipgloss.Center, lipgloss.Center, body),
	)
	return lipgloss.JoinVertical(lipgloss.Left, head, inner, foot)
}

// panelBox renders a titled bordered panel at the given content width.
func panelBox(title, content string, width int) string {
	if width < 10 {
		width = 10
	}
	inner := styleTitle.Render(title) + "\n" + content
	return stylePanel.Width(width).Render(inner)
}

// joinPanels stacks panels with a blank line, trimmed.
func joinPanels(parts ...string) string {
	var nonempty []string
	for _, p := range parts {
		if strings.TrimSpace(p) != "" {
			nonempty = append(nonempty, p)
		}
	}
	return lipgloss.JoinVertical(lipgloss.Center, nonempty...)
}
