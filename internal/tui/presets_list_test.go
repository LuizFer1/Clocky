package tui

import (
	"testing"

	"github.com/LuizFer1/Clocky/internal/presets"
)

func TestBuildPresetItems(t *testing.T) {
	s := &presets.Store{
		Pomodoros: []presets.PomodoroPreset{{Name: "Focus", Focus: 40, Break: 10, Long: 15, Cycles: 4, Auto: true}},
		Timers:    []presets.TimerPreset{{Name: "Break", Seconds: 300}},
	}
	items := buildPresetItems(s)
	if len(items) != 2 {
		t.Fatalf("len=%d", len(items))
	}
	if items[0].Kind != kindPomodoro || items[0].Name != "Focus" {
		t.Fatalf("item0=%+v", items[0])
	}
	if items[1].Kind != kindTimer || items[1].Name != "Break" {
		t.Fatalf("item1=%+v", items[1])
	}
	if items[0].Summary == "" || items[1].Summary == "" {
		t.Fatal("expected summaries")
	}
}
