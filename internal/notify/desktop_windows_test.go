//go:build windows

package notify

import (
	"testing"
)

func TestDesktopNotifyCmdHidesWindow(t *testing.T) {
	cmd := desktopNotifyCmd("Title", "Body")
	if cmd.SysProcAttr == nil {
		t.Fatal("expected SysProcAttr")
	}
	if !cmd.SysProcAttr.HideWindow {
		t.Fatal("HideWindow should be true")
	}
	if cmd.SysProcAttr.CreationFlags&createNoWindow == 0 {
		t.Fatalf("CreationFlags=%#x want CREATE_NO_WINDOW %#x", cmd.SysProcAttr.CreationFlags, createNoWindow)
	}
	joined := false
	for _, a := range cmd.Args {
		if a == "Hidden" {
			joined = true
		}
	}
	if !joined {
		t.Fatalf("args=%v want -WindowStyle Hidden", cmd.Args)
	}
}
