package cli

import "fmt"

func Run(args []string) error {
	if len(args) == 0 {
		printHelp()
		return nil
	}
	switch args[0] {
	case "help", "-h", "--help":
		printHelp()
		return nil
	default:
		return fmt.Errorf("unknown command %q\nrun: clocky help", args[0])
	}
}

func printHelp() {
	fmt.Print(`Clocky — terminal time manager

Usage:
  clocky pomodoro [name] [flags]
  clocky timer <duration|name>
  clocky timer --stop
  clocky stopwatch
  clocky add pomodoro --name <name> [flags]
  clocky add timer --name <name> <duration>
  clocky list [pomodoro|timer]
  clocky remove <pomodoro|timer> <name>
  clocky status
  clocky help
`)
}
