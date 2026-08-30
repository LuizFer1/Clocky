package presets

import (
	"testing"
)

func TestLoadMissingReturnsEmpty(t *testing.T) {
	root := t.TempDir()
	s, err := Load(root)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(s.Pomodoros) != 0 || len(s.Timers) != 0 {
		t.Fatalf("expected empty store, got %+v", s)
	}
}

func TestPomodoroUpsertGetOverwriteRemoveRoundTrip(t *testing.T) {
	root := t.TempDir()
	s, err := Load(root)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	p := PomodoroPreset{Name: "deep", Focus: 50, Break: 10, Long: 20, Cycles: 3, Auto: true}
	s.UpsertPomodoro(p)
	if err := s.Save(root); err != nil {
		t.Fatalf("Save: %v", err)
	}

	loaded, err := Load(root)
	if err != nil {
		t.Fatalf("Load after save: %v", err)
	}
	got, ok := loaded.GetPomodoro("deep")
	if !ok {
		t.Fatal("GetPomodoro: expected found")
	}
	if got != p {
		t.Fatalf("GetPomodoro = %+v want %+v", got, p)
	}
	if len(loaded.Pomodoros) != 1 {
		t.Fatalf("list len = %d want 1", len(loaded.Pomodoros))
	}

	updated := PomodoroPreset{Name: "deep", Focus: 25, Break: 5, Long: 15, Cycles: 4, Auto: false}
	loaded.UpsertPomodoro(updated)
	got, ok = loaded.GetPomodoro("deep")
	if !ok || got != updated {
		t.Fatalf("overwrite GetPomodoro = %+v ok=%v want %+v", got, ok, updated)
	}
	if len(loaded.Pomodoros) != 1 {
		t.Fatalf("after overwrite list len = %d want 1", len(loaded.Pomodoros))
	}
	if err := loaded.Save(root); err != nil {
		t.Fatalf("Save overwrite: %v", err)
	}

	loaded, err = Load(root)
	if err != nil {
		t.Fatalf("Load after overwrite: %v", err)
	}
	if err := loaded.RemovePomodoro("deep"); err != nil {
		t.Fatalf("RemovePomodoro: %v", err)
	}
	if _, ok := loaded.GetPomodoro("deep"); ok {
		t.Fatal("expected missing after remove")
	}
	if err := loaded.RemovePomodoro("deep"); err == nil {
		t.Fatal("RemovePomodoro missing: expected error")
	}
	if err := loaded.Save(root); err != nil {
		t.Fatalf("Save after remove: %v", err)
	}

	loaded, err = Load(root)
	if err != nil {
		t.Fatalf("final Load: %v", err)
	}
	if len(loaded.Pomodoros) != 0 {
		t.Fatalf("final list len = %d want 0", len(loaded.Pomodoros))
	}
}

func TestTimerUpsertGetOverwriteRemoveRoundTrip(t *testing.T) {
	root := t.TempDir()
	s, err := Load(root)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	tp := TimerPreset{Name: "tea", Seconds: 180}
	s.UpsertTimer(tp)
	if err := s.Save(root); err != nil {
		t.Fatalf("Save: %v", err)
	}

	loaded, err := Load(root)
	if err != nil {
		t.Fatalf("Load after save: %v", err)
	}
	got, ok := loaded.GetTimer("tea")
	if !ok {
		t.Fatal("GetTimer: expected found")
	}
	if got != tp {
		t.Fatalf("GetTimer = %+v want %+v", got, tp)
	}
	if len(loaded.Timers) != 1 {
		t.Fatalf("list len = %d want 1", len(loaded.Timers))
	}

	updated := TimerPreset{Name: "tea", Seconds: 240}
	loaded.UpsertTimer(updated)
	got, ok = loaded.GetTimer("tea")
	if !ok || got != updated {
		t.Fatalf("overwrite GetTimer = %+v ok=%v want %+v", got, ok, updated)
	}
	if len(loaded.Timers) != 1 {
		t.Fatalf("after overwrite list len = %d want 1", len(loaded.Timers))
	}
	if err := loaded.Save(root); err != nil {
		t.Fatalf("Save overwrite: %v", err)
	}

	loaded, err = Load(root)
	if err != nil {
		t.Fatalf("Load after overwrite: %v", err)
	}
	if err := loaded.RemoveTimer("tea"); err != nil {
		t.Fatalf("RemoveTimer: %v", err)
	}
	if _, ok := loaded.GetTimer("tea"); ok {
		t.Fatal("expected missing after remove")
	}
	if err := loaded.RemoveTimer("tea"); err == nil {
		t.Fatal("RemoveTimer missing: expected error")
	}
	if err := loaded.Save(root); err != nil {
		t.Fatalf("Save after remove: %v", err)
	}

	loaded, err = Load(root)
	if err != nil {
		t.Fatalf("final Load: %v", err)
	}
	if len(loaded.Timers) != 0 {
		t.Fatalf("final list len = %d want 0", len(loaded.Timers))
	}
}
