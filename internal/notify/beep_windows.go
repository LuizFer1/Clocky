//go:build windows

package notify

import "syscall"

var (
	kernel32 = syscall.NewLazyDLL("kernel32.dll")
	procBeep = kernel32.NewProc("Beep")
)

// nativeBeep plays a short ascending alarm chirp (Hz, milliseconds).
func nativeBeep() {
	beepTone(880, 110)
	beepTone(1175, 110)
	beepTone(1568, 160)
}

func beepTone(freq, ms uint32) {
	_, _, _ = procBeep.Call(uintptr(freq), uintptr(ms))
}
