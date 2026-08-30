//go:build !windows

package notify

func nativeAlert() {
	_ = Default{}.Beep()
}

func nativeStopAlert() {}
