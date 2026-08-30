package version

import "testing"

func TestStringDefaultDev(t *testing.T) {
	// Reset in case other tests mutate (use local assignment pattern in String tests)
	if got := format("dev", "", ""); got != "dev" {
		t.Fatalf("got %q", got)
	}
	if got := format("1.2.0", "abc1234", "2026-08-29"); got != "1.2.0 (abc1234 2026-08-29)" {
		t.Fatalf("got %q", got)
	}
}
