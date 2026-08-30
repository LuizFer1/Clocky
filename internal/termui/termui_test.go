package termui

import (
	"bytes"
	"strings"
	"testing"
)

func TestClearWritesANSIToBuffer(t *testing.T) {
	var buf bytes.Buffer
	if err := Clear(&buf); err != nil {
		t.Fatal(err)
	}
	got := buf.String()
	if !strings.Contains(got, "\033[2J") || !strings.Contains(got, "\033[H") {
		t.Fatalf("expected ANSI clear sequences, got %q", got)
	}
}

func TestEnableANSIDoesNotPanic(t *testing.T) {
	EnableANSI()
}
