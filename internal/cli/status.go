package cli

import (
	"fmt"
	"time"

	"github.com/luisf/clocky/internal/duration"
	"github.com/luisf/clocky/internal/stopwatch"
	"github.com/luisf/clocky/internal/timer"
)

func runStatus(args []string) error {
	if len(args) > 0 {
		return fmt.Errorf("usage: clocky status")
	}
	root, err := rootDir()
	if err != nil {
		return err
	}
	now := time.Now()

	swRunning, swElapsed, err := stopwatch.Status(root, now)
	if err != nil {
		return err
	}
	if swRunning {
		fmt.Printf("Stopwatch: running (%s)\n", duration.Format(swElapsed))
	} else {
		fmt.Println("Stopwatch: idle")
	}

	tmActive, tmRemaining, tmLabel, err := timer.Status(root, now)
	if err != nil {
		return err
	}
	if tmActive {
		label := tmLabel
		if label == "" {
			label = "(unnamed)"
		}
		fmt.Printf("Timer: active %q remaining %s\n", label, duration.Format(tmRemaining))
	} else {
		fmt.Println("Timer: idle")
	}
	return nil
}
