package notify

import (
	"fmt"
	"io"
	"os"
)

// Notifier delivers completion / phase-change alerts.
type Notifier interface {
	Beep() error
	Banner(title, body string) error
	Desktop(title, body string) error
}

// Default writes beep/banner to Out (stdout if nil) and best-effort native toasts.
type Default struct {
	Out io.Writer
}

func (d Default) out() io.Writer {
	if d.Out != nil {
		return d.Out
	}
	return os.Stdout
}

func (d Default) Beep() error {
	_, err := fmt.Fprint(d.out(), "\a")
	return err
}

func (d Default) Banner(title, body string) error {
	_, err := fmt.Fprintf(d.out(), "\n*** %s ***\n%s\n\n", title, body)
	return err
}

func (d Default) Desktop(title, body string) error {
	return desktopNotify(title, body)
}

// All runs Beep, Banner, then Desktop in sequence.
func All(n Notifier, title, body string) error {
	if err := n.Beep(); err != nil {
		return err
	}
	if err := n.Banner(title, body); err != nil {
		return err
	}
	return n.Desktop(title, body)
}
