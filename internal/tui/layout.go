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
	return lipgloss.Place(w, h, lipgloss.Center, lipgloss.Center, content)
}

// fillFrame paints a full-terminal frame: CLOCKY banner, breadcrumb, body, footer.
func fillFrame(w, h int, header, body, footer string) string {
	if w < 1 {
		w = 80
	}
	if h < 1 {
		h = 24
	}

	banner := clockyBanner(w)
	sub := ""
	if strings.TrimSpace(header) != "" {
		crumb := styleDim.Render("clocky › ") + stylePhase.Render(header)
		sub = lipgloss.PlaceHorizontal(w, lipgloss.Center, crumb)
	}
	foot := renderFooter(footer, w)

	used := lipgloss.Height(banner) + lipgloss.Height(foot)
	if sub != "" {
		used += lipgloss.Height(sub)
	}
	bodyH := h - used
	if bodyH < 1 {
		bodyH = 1
	}

	inner := lipgloss.NewStyle().Width(w).Height(bodyH).Foreground(colText).Render(
		lipgloss.Place(w, bodyH, lipgloss.Center, lipgloss.Center, body),
	)
	parts := []string{banner}
	if sub != "" {
		parts = append(parts, sub)
	}
	parts = append(parts, inner, foot)
	return lipgloss.JoinVertical(lipgloss.Left, parts...)
}

// renderFooter turns "key desc  key desc" pairs (two-space separated) into
// a status bar with highlighted keys.
func renderFooter(footer string, w int) string {
	var parts []string
	for _, chunk := range strings.Split(footer, "  ") {
		chunk = strings.TrimSpace(chunk)
		if chunk == "" {
			continue
		}
		key, desc, ok := strings.Cut(chunk, " ")
		if !ok {
			parts = append(parts, styleMuted.Render(chunk))
			continue
		}
		parts = append(parts, styleKey.Render(key)+" "+styleMuted.Render(desc))
	}
	line := strings.Join(parts, styleDim.Render("  ·  "))
	return lipgloss.NewStyle().Width(w).Padding(0, 2).Background(colSurface).Render(
		lipgloss.PlaceHorizontal(max(1, w-4), lipgloss.Center, line),
	)
}

// panelBox renders a titled bordered panel at the given content width.
func panelBox(title, content string, width int) string {
	return panelBoxFocused(title, content, width, false)
}

// panelBoxFocused is panelBox with a highlighted border when focused.
func panelBoxFocused(title, content string, width int, focused bool) string {
	if width < 10 {
		width = 10
	}
	st := stylePanel
	head := styleTitle.Render(title)
	if focused {
		st = stylePanelFocused
		head = styleTitle.Render("▸ " + title)
	}
	return st.Width(width).Render(head + "\n" + content)
}

// joinPanels stacks panels, trimmed of empties.
func joinPanels(parts ...string) string {
	var nonempty []string
	for _, p := range parts {
		if strings.TrimSpace(p) != "" {
			nonempty = append(nonempty, p)
		}
	}
	return lipgloss.JoinVertical(lipgloss.Center, nonempty...)
}
