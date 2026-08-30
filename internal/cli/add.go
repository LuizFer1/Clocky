package cli

import (
	"flag"
	"fmt"
	"time"

	"github.com/luisf/clocky/internal/duration"
	"github.com/luisf/clocky/internal/presets"
)

func runAdd(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: clocky add <pomodoro|timer> ...")
	}
	switch args[0] {
	case "pomodoro":
		return addPomodoro(args[1:])
	case "timer":
		return addTimer(args[1:])
	default:
		return fmt.Errorf("usage: clocky add <pomodoro|timer> ...")
	}
}

func addPomodoro(args []string) error {
	fs := flag.NewFlagSet("add pomodoro", flag.ContinueOnError)
	name := fs.String("name", "", "preset name")
	focus := fs.Int("focus", 25, "focus duration in minutes")
	brk := fs.Int("break", 5, "short break in minutes")
	long := fs.Int("long", 15, "long break in minutes")
	cycles := fs.Int("cycles", 4, "focus sessions before long break")
	auto := fs.Bool("auto", false, "advance phases without pressing Enter")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *name == "" {
		return fmt.Errorf("usage: clocky add pomodoro --name <name> [flags]")
	}
	if len(fs.Args()) > 0 {
		return fmt.Errorf("usage: clocky add pomodoro --name <name> [flags]")
	}
	if *cycles < 1 {
		return fmt.Errorf("--cycles must be >= 1")
	}

	root, err := rootDir()
	if err != nil {
		return err
	}
	store, err := presets.Load(root)
	if err != nil {
		return err
	}
	_, existed := store.GetPomodoro(*name)
	store.UpsertPomodoro(presets.PomodoroPreset{
		Name:   *name,
		Focus:  *focus,
		Break:  *brk,
		Long:   *long,
		Cycles: *cycles,
		Auto:   *auto,
	})
	if err := store.Save(root); err != nil {
		return err
	}
	if existed {
		fmt.Printf("Updated pomodoro preset %q\n", *name)
	} else {
		fmt.Printf("Saved pomodoro preset %q\n", *name)
	}
	return nil
}

func addTimer(args []string) error {
	fs := flag.NewFlagSet("add timer", flag.ContinueOnError)
	name := fs.String("name", "", "preset name")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *name == "" || len(fs.Args()) != 1 {
		return fmt.Errorf("usage: clocky add timer --name <name> <duration>")
	}
	d, err := duration.Parse(fs.Args()[0])
	if err != nil {
		return fmt.Errorf("invalid duration: %v", err)
	}

	root, err := rootDir()
	if err != nil {
		return err
	}
	store, err := presets.Load(root)
	if err != nil {
		return err
	}
	_, existed := store.GetTimer(*name)
	store.UpsertTimer(presets.TimerPreset{
		Name:    *name,
		Seconds: int64(d / time.Second),
	})
	if err := store.Save(root); err != nil {
		return err
	}
	if existed {
		fmt.Printf("Updated timer preset %q (%s)\n", *name, duration.Format(d))
	} else {
		fmt.Printf("Saved timer preset %q (%s)\n", *name, duration.Format(d))
	}
	return nil
}
