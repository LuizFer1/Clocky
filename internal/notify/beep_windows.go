//go:build windows

package notify

import "syscall"

var (
	user32           = syscall.NewLazyDLL("user32.dll")
	procMessageBeep  = user32.NewProc("MessageBeep")
	mbIconAsterisk   = uintptr(0x00000040)
)

// nativeBeep uses MessageBeep so audio still plays when the terminal ignores
// ASCII BEL (common with some Windows hosts / alt screens).
func nativeBeep() {
	_, _, _ = procMessageBeep.Call(mbIconAsterisk)
}
