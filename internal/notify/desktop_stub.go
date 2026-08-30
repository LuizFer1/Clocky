//go:build !windows && !darwin && !linux

package notify

func desktopNotify(title, body string) error {
	return nil
}
