package cli

import (
	"flag"
	"fmt"
	"time"

	"github.com/luisf/clocky/internal/pomodoro"
	"github.com/luisf/clocky/internal/presets"
)

func runPomodoro(args []string) error {
	fs := flag.NewFlagSet("pomodoro", flag.ContinueOnError)
	focus := fs.Int("focus", 25, "focus duration in minutes")
	brk := fs.Int("break", 5, "short break in minutes")
	long := fs.Int("long", 15, "long break in minutes")
	cycles := fs.Int("cycles", 4, "focus sessions before long break")
	auto := fs.Bool("auto", false, "advance phases without pressing Enter")
	if err := fs.Parse(args); err != nil {
		return err
	}

	cfg := pomodoro.Config{
		Focus:  time.Duration(*focus) * time.Minute,
		Break:  time.Duration(*brk) * time.Minute,
		Long:   time.Duration(*long) * time.Minute,
		Cycles: *cycles,
		Auto:   *auto,
	}

	pos := fs.Args()
	if len(pos) > 1 {
		return fmt.Errorf("usage: clocky pomodoro [name] [flags]")
	}
	if len(pos) == 1 {
		root, err := rootDir()
		if err != nil {
			return err
		}
		store, err := presets.Load(root)
		if err != nil {
			return err
		}
		p, ok := store.GetPomodoro(pos[0])
		if !ok {
			return fmt.Errorf("pomodoro preset %q not found\nrun: clocky list", pos[0])
		}
		cfg = pomodoro.Config{
			Focus:  time.Duration(p.Focus) * time.Minute,
			Break:  time.Duration(p.Break) * time.Minute,
			Long:   time.Duration(p.Long) * time.Minute,
			Cycles: p.Cycles,
			Auto:   p.Auto,
		}
		fs.Visit(func(f *flag.Flag) {
			switch f.Name {
			case "focus":
				cfg.Focus = time.Duration(*focus) * time.Minute
			case "break":
				cfg.Break = time.Duration(*brk) * time.Minute
			case "long":
				cfg.Long = time.Duration(*long) * time.Minute
			case "cycles":
				cfg.Cycles = *cycles
			case "auto":
				cfg.Auto = *auto
			}
		})
	}

	if cfg.Cycles < 1 {
		return fmt.Errorf("--cycles must be >= 1")
	}
	return pomodoro.Run(cfg, pomodoro.Hooks{})
}
