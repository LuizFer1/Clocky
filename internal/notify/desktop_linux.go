//go:build linux

package notify

import "os/exec"

func desktopNotify(title, body string) error {
	path, err := exec.LookPath("notify-send")
	if err != nil {
		return nil
	}
	cmd := exec.Command(path, title, body)
	_ = cmd.Run()
	return nil
}
