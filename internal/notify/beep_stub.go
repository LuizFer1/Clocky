//go:build !windows

package notify

import "time"

func nativeBeep() {
	d := Default{}
	_ = d.Beep()
	time.Sleep(90 * time.Millisecond)
	_ = d.Beep()
	time.Sleep(90 * time.Millisecond)
	_ = d.Beep()
}
