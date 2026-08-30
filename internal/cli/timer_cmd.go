package cli

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/LuizFer1/Clocky/internal/duration"
	"github.com/LuizFer1/Clocky/internal/notify"
	"github.com/LuizFer1/Clocky/internal/presets"
	"github.com/LuizFer1/Clocky/internal/timer"
)

func runTimer(args []string) error {
	fs := flag.NewFlagSet("timer", flag.ContinueOnError)
	stop := fs.Bool("stop", false, "cancel the active timer")
	worker := fs.Bool("worker", false, "run background timer worker (internal)")
	if err := fs.Parse(args); err != nil {
		return err
	}

	root, err := rootDir()
	if err != nil {
		return err
	}

	if *stop {
		if *worker || len(fs.Args()) > 0 {
			return fmt.Errorf("usage: clocky timer --stop")
		}
		if err := timer.Stop(root); err != nil {
			return err
		}
		fmt.Println("Timer stopped")
		return nil
	}

	if *worker {
		if *stop || len(fs.Args()) > 0 {
			return fmt.Errorf("usage: clocky timer --worker")
		}
		return timer.Worker(root, notify.Default{}, nil, nil)
	}

	if len(fs.Args()) != 1 {
		return fmt.Errorf("usage: clocky timer <H:M:S|name>")
	}
	arg := fs.Args()[0]

	d, label, err := resolveTimer(root, arg)
	if err != nil {
		return err
	}

	exe, err := os.Executable()
	if err != nil {
		return err
	}
	if err := timer.Start(root, d, label, exe, []string{"timer", "--worker"}); err != nil {
		if strings.Contains(err.Error(), "timer already running") {
			return fmt.Errorf("timer already running\nrun: clocky status  or  clocky timer --stop")
		}
		return err
	}
	fmt.Printf("Timer started: %s (%s)\n", label, duration.Format(d))
	return nil
}

func resolveTimer(root, arg string) (time.Duration, string, error) {
	if d, err := duration.Parse(arg); err == nil {
		return d, arg, nil
	}
	store, err := presets.Load(root)
	if err != nil {
		return 0, "", err
	}
	t, ok := store.GetTimer(arg)
	if !ok {
		return 0, "", fmt.Errorf("unknown timer %q\nusage: clocky timer <H:M:S|name>\nrun: clocky list", arg)
	}
	return time.Duration(t.Seconds) * time.Second, t.Name, nil
}
