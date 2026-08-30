//go:build windows

package notify

import (
	"fmt"
	"os/exec"
	"strings"
)

// desktopNotify shows a Windows toast via PowerShell. Failures are ignored
// so timers/pomodoros still complete when toast APIs are unavailable.
func desktopNotify(title, body string) error {
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

	cmd := exec.Command("powershell", "-NoProfile", "-NonInteractive", "-Command", script)
	_ = cmd.Run()
	return nil
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
