//go:build !windows

package notify

import "time"

func nativeAlert() {
	d := Default{}
	_ = d.Beep()
	time.Sleep(80 * time.Millisecond)
	_ = d.Beep()
}

func nativeStopAlert() {}
