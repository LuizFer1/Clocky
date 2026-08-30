// Package termui helps redraw live terminal UIs (pomodoro clock).
package termui

import "io"

// Clear clears the visible screen and moves the cursor home when possible.
// On plain Writers (buffers, pipes) it emits ANSI clear sequences.
func Clear(w io.Writer) error {
	if clearNative(w) {
		return nil
	}
	_, err := io.WriteString(w, "\033[2J\033[H")
	return err
}

// EnableANSI turns on Windows virtual-terminal processing for stdout/stderr
// when attached to a console. No-op on other platforms.
func EnableANSI() {
	enableANSI()
}
