package notify

// Alert plays an audible cue for the interactive terminal (TUI hub).
// Uses the ASCII bell plus a platform native beep when available.
func Alert() {
	_ = Default{}.Beep()
	nativeBeep()
	_ = Default{}.Beep()
}
