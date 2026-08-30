//go:build windows

package notify

import (
	"os"
	"path/filepath"
	"sync"
	"syscall"
	"unsafe"
)

const (
	sndAsync     = 0x0001
	sndNoDefault = 0x0002
	sndFilename  = 0x00020000
	sndNoStop    = 0x0010
)

var (
	winmm          = syscall.NewLazyDLL("winmm.dll")
	procPlaySoundW = winmm.NewProc("PlaySoundW")
	alertMu        sync.Mutex
)

// Prefer short, clean system sounds over kernel32 Beep (often distorted).
var alertSoundCandidates = []string{
	"Windows Notify System Generic.wav",
	"Windows Notify Calendar.wav",
	"notify.wav",
	"Windows Background.wav",
	"ding.wav",
	"Alarm01.wav",
}

func nativeAlert() {
	alertMu.Lock()
	defer alertMu.Unlock()
	nativeStopAlertLocked()
	path := firstExistingAlertSound()
	if path == "" {
		// Soft fallback: single MessageBeep, never the harsh multi-tone Beep API.
		messageBeep()
		return
	}
	playSoundW(path, sndAsync|sndFilename|sndNoDefault)
}

func nativeStopAlert() {
	alertMu.Lock()
	defer alertMu.Unlock()
	nativeStopAlertLocked()
}

func nativeStopAlertLocked() {
	playSoundW("", 0)
}

func firstExistingAlertSound() string {
	media := filepath.Join(os.Getenv("WINDIR"), "Media")
	for _, name := range alertSoundCandidates {
		p := filepath.Join(media, name)
		if st, err := os.Stat(p); err == nil && !st.IsDir() {
			return p
		}
	}
	return ""
}

func playSoundW(path string, flags uintptr) {
	var name *uint16
	if path != "" {
		name, _ = syscall.UTF16PtrFromString(path)
	}
	_, _, _ = procPlaySoundW.Call(
		uintptr(unsafe.Pointer(name)),
		0,
		flags,
	)
}

func messageBeep() {
	user32 := syscall.NewLazyDLL("user32.dll")
	proc := user32.NewProc("MessageBeep")
	_, _, _ = proc.Call(0x00000040) // MB_ICONASTERISK
}
