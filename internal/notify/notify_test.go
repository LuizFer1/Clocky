package notify

import (
	"bytes"
	"strings"
	"testing"
)

type RecordingNotifier struct {
	Events []string
}

func (r *RecordingNotifier) Beep() error {
	r.Events = append(r.Events, "beep")
	return nil
}

func (r *RecordingNotifier) Banner(title, body string) error {
	r.Events = append(r.Events, "banner:"+title+":"+body)
	return nil
}

func (r *RecordingNotifier) Desktop(title, body string) error {
	r.Events = append(r.Events, "desktop:"+title+":"+body)
	return nil
}

func TestDefaultBannerWritesExpectedText(t *testing.T) {
	var buf bytes.Buffer
	d := Default{Out: &buf}
	if err := d.Banner("Timer done", "tea finished"); err != nil {
		t.Fatalf("Banner: %v", err)
	}
	got := buf.String()
	want := "\n*** Timer done ***\ntea finished\n\n"
	if got != want {
		t.Fatalf("Banner output = %q want %q", got, want)
	}
}

func TestDefaultBeepWritesBEL(t *testing.T) {
	var buf bytes.Buffer
	d := Default{Out: &buf}
	if err := d.Beep(); err != nil {
		t.Fatalf("Beep: %v", err)
	}
	if got := buf.String(); got != "\a" {
		t.Fatalf("Beep output = %q want BEL", got)
	}
}

func TestAllCallsBeepBannerDesktop(t *testing.T) {
	r := &RecordingNotifier{}
	if err := All(r, "Title", "Body"); err != nil {
		t.Fatalf("All: %v", err)
	}
	if len(r.Events) != 3 {
		t.Fatalf("events = %v want 3", r.Events)
	}
	if r.Events[0] != "beep" {
		t.Fatalf("first = %q want beep", r.Events[0])
	}
	if !strings.HasPrefix(r.Events[1], "banner:") {
		t.Fatalf("second = %q want banner", r.Events[1])
	}
	if !strings.HasPrefix(r.Events[2], "desktop:") {
		t.Fatalf("third = %q want desktop", r.Events[2])
	}
}
