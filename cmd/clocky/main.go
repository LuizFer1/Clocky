package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/LuizFer1/Clocky/internal/cli"
	"github.com/LuizFer1/Clocky/internal/termui"
	"github.com/LuizFer1/Clocky/internal/update"
)

func main() {
	termui.EnableANSI()
	cleanupOldBinary()
	if err := cli.Run(os.Args[1:]); err != nil {
		var ec *cli.ExitCodeError
		if errors.As(err, &ec) {
			if ec.Msg != "" {
				fmt.Fprintln(os.Stderr, ec.Msg)
			}
			os.Exit(ec.Code)
		}
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func cleanupOldBinary() {
	exe, err := os.Executable()
	if err != nil {
		return
	}
	update.CleanupStaleOld(exe)
}
