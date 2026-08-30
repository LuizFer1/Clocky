package cli

import (
	"fmt"

	"github.com/LuizFer1/Clocky/internal/version"
)

func runVersion(args []string) error {
	if len(args) > 0 {
		return fmt.Errorf("usage: clocky version")
	}
	fmt.Println(version.String())
	return nil
}
