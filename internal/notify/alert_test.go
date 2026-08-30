package notify

import "testing"

func TestAlertDoesNotPanic(t *testing.T) {
	Alert() // best-effort audio; must not crash in CI
}
