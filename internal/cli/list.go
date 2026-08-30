package cli

import (
	"fmt"
	"time"

	"github.com/luisf/clocky/internal/duration"
	"github.com/luisf/clocky/internal/presets"
)

func runList(args []string) error {
	kind := ""
	if len(args) > 1 {
		return fmt.Errorf("usage: clocky list [pomodoro|timer]")
	}
	if len(args) == 1 {
		switch args[0] {
		case "pomodoro", "timer":
			kind = args[0]
		default:
			return fmt.Errorf("usage: clocky list [pomodoro|timer]")
		}
	}

	root, err := rootDir()
	if err != nil {
		return err
	}
	store, err := presets.Load(root)
	if err != nil {
		return err
	}

	showPomo := kind == "" || kind == "pomodoro"
	showTimer := kind == "" || kind == "timer"

	if showPomo {
		fmt.Println("Pomodoros:")
		if len(store.Pomodoros) == 0 {
			fmt.Println("  (none)")
		} else {
			for _, p := range store.Pomodoros {
				fmt.Printf("  %s  focus=%d break=%d long=%d cycles=%d auto=%v\n",
					p.Name, p.Focus, p.Break, p.Long, p.Cycles, p.Auto)
			}
		}
	}
	if showTimer {
		if showPomo {
			fmt.Println()
		}
		fmt.Println("Timers:")
		if len(store.Timers) == 0 {
			fmt.Println("  (none)")
		} else {
			for _, t := range store.Timers {
				d := time.Duration(t.Seconds) * time.Second
				fmt.Printf("  %s  %s\n", t.Name, duration.Format(d))
			}
		}
	}
	return nil
}
