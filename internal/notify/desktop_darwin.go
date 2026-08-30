//go:build darwin

package notify

import (
	"fmt"
	"os/exec"
	"strings"
)

func desktopNotify(title, body string) error {
	script := fmt.Sprintf(
		`display notification %s with title %s`,
		osaQuote(body),
		osaQuote(title),
	)
	cmd := exec.Command("osascript", "-e", script)
	_ = cmd.Run()
	return nil
}

func osaQuote(s string) string {
	return `"` + strings.ReplaceAll(s, `"`, `\"`) + `"`
}
