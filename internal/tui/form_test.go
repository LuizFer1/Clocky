package tui

import (
	"testing"

	"github.com/LuizFer1/Clocky/internal/presets"
)

func TestSavePomodoroPreset(t *testing.T) {
	root := t.TempDir()
	err := savePomodoroPreset(root, presets.PomodoroPreset{
		Name: "X", Focus: 25, Break: 5, Long: 15, Cycles: 4,
	})
	if err != nil {
		t.Fatal(err)
	}
	st, err := presets.Load(root)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := st.GetPomodoro("X"); !ok {
		t.Fatal("missing preset")
	}
}

func TestSaveTimerPreset(t *testing.T) {
	root := t.TempDir()
	err := saveTimerPreset(root, presets.TimerPreset{Name: "Break", Seconds: 300})
	if err != nil {
		t.Fatal(err)
	}
	st, err := presets.Load(root)
	if err != nil {
		t.Fatal(err)
	}
	got, ok := st.GetTimer("Break")
	if !ok || got.Seconds != 300 {
		t.Fatalf("got=%+v ok=%v", got, ok)
	}
}

func TestDeletePreset(t *testing.T) {
	root := t.TempDir()
	_ = savePomodoroPreset(root, presets.PomodoroPreset{Name: "X", Focus: 25, Break: 5, Long: 15, Cycles: 1})
	if err := deletePreset(root, kindPomodoro, "X"); err != nil {
		t.Fatal(err)
	}
	st, _ := presets.Load(root)
	if _, ok := st.GetPomodoro("X"); ok {
		t.Fatal("still present")
	}
}
