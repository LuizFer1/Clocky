//go:build windows

package termui

import (
	"io"

	"golang.org/x/sys/windows"
)

const enableVirtualTerminalProcessing = 0x0004

func enableANSI() {
	enableVT(windows.Stdout)
	enableVT(windows.Stderr)
}

func enableVT(handle windows.Handle) {
	var mode uint32
	if err := windows.GetConsoleMode(handle, &mode); err != nil {
		return
	}
	_ = windows.SetConsoleMode(handle, mode|enableVirtualTerminalProcessing)
}

// clearNative is unused on Windows once VT is enabled; ANSI clear is preferred.
func clearNative(w io.Writer) bool {
	_ = w
	return false
}
