//go:build windows

package notify

import "syscall"

const (
	mbIconAsterisk     = uintptr(0x00000040)
	mbIconExclamation  = uintptr(0x00000030)
	mbIconHand         = uintptr(0x00000010)
	mbSimpleBeep       = uintptr(0xFFFFFFFF)
)

var (
	user32          = syscall.NewLazyDLL("user32.dll")
	procMessageBeep  = user32.NewProc("MessageBeep")
)

// nativeAlert uses the Windows MessageBeep exclamation tone — cleaner than
// kernel32 Beep frequency sweeps on modern machines.
func nativeAlert() {
	_, _, _ = procMessageBeep.Call(mbIconExclamation)
}

func nativeStopAlert() {
	// MessageBeep is instantaneous; nothing to cancel.
}
