package notify

// Alert plays one alarm pulse for the TUI hub (system MessageBeep on Windows).
func Alert() {
	nativeAlert()
}

// StopAlert stops any in-progress hub alarm sound when applicable.
func StopAlert() {
	nativeStopAlert()
}
