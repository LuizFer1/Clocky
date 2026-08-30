package duration

import (
	"testing"
	"time"
)

func TestParse(t *testing.T) {
	cases := []struct {
		in   string
		want time.Duration
	}{
		{"1:30:00", time.Hour + 30*time.Minute},
		{"1:30:", time.Hour + 30*time.Minute},
		{"1::30", time.Hour + 30*time.Second},
		{":5:30", 5*time.Minute + 30*time.Second},
		{"24::", 24 * time.Hour},
		{":25:", 25 * time.Minute},
		{"::50", 50 * time.Second},
		{":1440:", 24 * time.Hour},
		{"::1440", 24 * time.Minute},
		{":90:00", 90 * time.Minute},
		{"90", 90 * time.Second},
	}
	for _, tc := range cases {
		got, err := Parse(tc.in)
		if err != nil {
			t.Fatalf("Parse(%q): %v", tc.in, err)
		}
		if got != tc.want {
			t.Fatalf("Parse(%q)=%v want %v", tc.in, got, tc.want)
		}
	}
}

func TestParseInvalid(t *testing.T) {
	for _, in := range []string{"", "abc", "1:2:3:4", "::", ":"} {
		if _, err := Parse(in); err == nil {
			t.Fatalf("Parse(%q) expected error", in)
		}
	}
}

func TestFormat(t *testing.T) {
	if got := Format(time.Hour + 2*time.Minute + 3*time.Second); got != "01:02:03" {
		t.Fatalf("got %q", got)
	}
	if got := Format(24 * time.Hour); got != "24:00:00" {
		t.Fatalf("got %q", got)
	}
}
