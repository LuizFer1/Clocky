package notify

// Alert plays one alarm pulse for the TUI hub.
func Alert() {
	nativeAlert()
}

// StopAlert stops any in-progress hub alarm sound.
func StopAlert() {
	nativeStopAlert()
}
