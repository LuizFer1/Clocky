package cli

import (
	"flag"
	"os"
	"runtime"

	"github.com/luisf/clocky/internal/update"
	"github.com/luisf/clocky/internal/version"
)

func runUpdate(args []string) error {
	fs := flag.NewFlagSet("update", flag.ContinueOnError)
	yes := fs.Bool("yes", false, "update without prompting")
	check := fs.Bool("check", false, "only check for updates")
	if err := fs.Parse(args); err != nil {
		return err
	}
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	res, err := update.Run(update.Options{
		CurrentVersion: version.Version,
		GOOS:           runtime.GOOS,
		GOARCH:         runtime.GOARCH,
		ExePath:        exe,
		Yes:            *yes,
		CheckOnly:      *check,
		Stdin:          os.Stdin,
		Stdout:         os.Stdout,
	})
	if err != nil {
		return err
	}
	if *check && res.Pending {
		return &ExitCodeError{Code: 3, Msg: ""} // message already printed to Stdout
	}
	return nil
}
