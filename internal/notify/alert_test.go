package notify

import "testing"

func TestAlertDoesNotPanic(t *testing.T) {
	Alert()
	StopAlert()
}
