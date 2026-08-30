package cli

import (
	"os"

	"golang.org/x/term"
)

func isInteractive() bool {
	return term.IsTerminal(int(os.Stdout.Fd())) && term.IsTerminal(int(os.Stdin.Fd()))
}
