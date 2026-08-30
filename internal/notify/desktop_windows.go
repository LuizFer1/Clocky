//go:build windows

package notify

import (
	"fmt"
	"os/exec"
	"strings"
	"syscall"
)

// CREATE_NO_WINDOW prevents a console window from appearing when spawning
// powershell.exe for toast notifications.
const createNoWindow = 0x08000000

// desktopNotify shows a Windows toast via PowerShell. Failures are ignored
// so timers/pomodoros still complete when toast APIs are unavailable.
func desktopNotify(title, body string) error {
	cmd := desktopNotifyCmd(title, body)
	_ = cmd.Run()
	return nil
}

func desktopNotifyCmd(title, body string) *exec.Cmd {
	script := fmt.Sprintf(`
$ErrorActionPreference = 'Stop'
try {
  [Windows.UI.Notifications.ToastNotificationManager, Windows.UI.Notifications, ContentType = WindowsRuntime] | Out-Null
  [Windows.Data.Xml.Dom.XmlDocument, Windows.Data.Xml.Dom.XmlDocument, ContentType = WindowsRuntime] | Out-Null
  $xml = New-Object Windows.Data.Xml.Dom.XmlDocument
  $xml.LoadXml(@'
<toast>
  <visual>
    <binding template="ToastText02">
      <text id="1">%s</text>
      <text id="2">%s</text>
    </binding>
  </visual>
</toast>
'@)
  $toast = [Windows.UI.Notifications.ToastNotification]::new($xml)
  [Windows.UI.Notifications.ToastNotificationManager]::CreateToastNotifier('Clocky').Show($toast)
} catch {
  exit 0
}
`, xmlEscape(title), xmlEscape(body))

	cmd := exec.Command(
		"powershell.exe",
		"-NoProfile",
		"-NonInteractive",
		"-WindowStyle", "Hidden",
		"-Command", script,
	)
	cmd.Stdin = nil
	cmd.Stdout = nil
	cmd.Stderr = nil
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow:    true,
		CreationFlags: createNoWindow,
	}
	return cmd
}

func xmlEscape(s string) string {
	r := strings.NewReplacer(
		"&", "&amp;",
		"<", "&lt;",
		">", "&gt;",
		`"`, "&quot;",
		"'", "&apos;",
	)
	return r.Replace(s)
}
