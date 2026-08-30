package cli

import (
	"fmt"
	"time"

	"github.com/LuizFer1/Clocky/internal/duration"
	"github.com/LuizFer1/Clocky/internal/stopwatch"
)

func runStopwatch(args []string) error {
	if len(args) > 0 {
		return fmt.Errorf("usage: clocky stopwatch")
	}
	root, err := rootDir()
	if err != nil {
		return err
	}
	action, elapsed, err := stopwatch.Toggle(root, time.Now())
	if err != nil {
		return err
	}
	switch action {
	case "started":
		fmt.Println("Stopwatch started")
	case "stopped":
		fmt.Printf("Stopwatch stopped: %s\n", duration.Format(elapsed))
	default:
		return fmt.Errorf("unexpected stopwatch action %q", action)
	}
	return nil
}
