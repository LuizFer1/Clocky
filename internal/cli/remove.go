package cli

import (
	"fmt"

	"github.com/LuizFer1/Clocky/internal/presets"
)

func runRemove(args []string) error {
	if len(args) != 2 {
		return fmt.Errorf("usage: clocky remove <pomodoro|timer> <name>")
	}
	root, err := rootDir()
	if err != nil {
		return err
	}
	store, err := presets.Load(root)
	if err != nil {
		return err
	}

	switch args[0] {
	case "pomodoro":
		if err := store.RemovePomodoro(args[1]); err != nil {
			return err
		}
	case "timer":
		if err := store.RemoveTimer(args[1]); err != nil {
			return err
		}
	default:
		return fmt.Errorf("usage: clocky remove <pomodoro|timer> <name>")
	}
	if err := store.Save(root); err != nil {
		return err
	}
	fmt.Printf("Removed %s preset %q\n", args[0], args[1])
	return nil
}
