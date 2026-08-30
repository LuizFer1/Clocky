package cli

// ExitCodeError signals a non-1 process exit code to main.
type ExitCodeError struct {
	Code int
	Msg  string
}

func (e *ExitCodeError) Error() string {
	if e.Msg == "" {
		return ""
	}
	return e.Msg
}
