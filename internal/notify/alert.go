package notify

// Alert plays an audible cue for the interactive terminal (TUI hub).
// On Windows this is a multi-tone Beep sequence; elsewhere it falls back
// to repeated ASCII bells.
func Alert() {
	nativeBeep()
}
