package cli

import (
	"fmt"

	"github.com/LuizFer1/Clocky/internal/state"
	"github.com/LuizFer1/Clocky/internal/tui"
)

// Run dispatches clocky subcommands.
func Run(args []string) error {
	if len(args) == 0 {
		if isInteractive() {
			return tui.Run()
		}
		printHelp()
		return nil
	}
	switch args[0] {
	case "help", "-h", "--help":
		printHelp()
		return nil
	case "version":
		return runVersion(args[1:])
	case "update":
		return runUpdate(args[1:])
	case "pomodoro":
		return runPomodoro(args[1:])
	case "timer":
		return runTimer(args[1:])
	case "stopwatch":
		return runStopwatch(args[1:])
	case "add":
		return runAdd(args[1:])
	case "list":
		return runList(args[1:])
	case "remove":
		return runRemove(args[1:])
	case "status":
		return runStatus(args[1:])
	default:
		return fmt.Errorf("unknown command %q\nrun: clocky help", args[0])
	}
}

func rootDir() (string, error) {
	root, err := state.DefaultRoot()
	if err != nil {
		return "", err
	}
	if err := state.EnsureDir(root); err != nil {
		return "", err
	}
	return root, nil
}

func printHelp() {
	fmt.Print(`Clocky — terminal time manager

Usage:
  clocky                 Open the Terminal UI (interactive terminal)
  clocky pomodoro [name] [flags]
  clocky timer <duration|name>
  clocky timer --stop
  clocky stopwatch
  clocky add pomodoro --name <name> [flags]
  clocky add timer --name <name> <duration>
  clocky list [pomodoro|timer]
  clocky remove <pomodoro|timer> <name>
  clocky status
  clocky version
  clocky update [--yes] [--check]
  clocky help

Pomodoro flags:
  --focus <min>   Focus duration in minutes (default 25)
  --break <min>   Short break in minutes (default 5)
  --long <min>    Long break in minutes (default 15)
  --cycles <n>    Focus sessions before long break (default 4)
  --auto          Advance phases without pressing Enter
`)
}
