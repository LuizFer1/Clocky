package tui

import (
	"fmt"
	"strings"

	"github.com/LuizFer1/Clocky/internal/presets"
)

func savePomodoroPreset(root string, p presets.PomodoroPreset) error {
	p.Name = strings.TrimSpace(p.Name)
	if p.Name == "" {
		return fmt.Errorf("name is required")
	}
	if p.Cycles < 1 {
		return fmt.Errorf("cycles must be >= 1")
	}
	if p.Focus <= 0 || p.Break <= 0 || p.Long <= 0 {
		return fmt.Errorf("focus, break, and long must be > 0")
	}
	st, err := presets.Load(root)
	if err != nil {
		return err
	}
	st.UpsertPomodoro(p)
	return st.Save(root)
}

func saveTimerPreset(root string, t presets.TimerPreset) error {
	t.Name = strings.TrimSpace(t.Name)
	if t.Name == "" {
		return fmt.Errorf("name is required")
	}
	if t.Seconds <= 0 {
		return fmt.Errorf("duration must be > 0")
	}
	st, err := presets.Load(root)
	if err != nil {
		return err
	}
	st.UpsertTimer(t)
	return st.Save(root)
}

func deletePreset(root, kind, name string) error {
	st, err := presets.Load(root)
	if err != nil {
		return err
	}
	switch kind {
	case kindPomodoro:
		if err := st.RemovePomodoro(name); err != nil {
			return err
		}
	case kindTimer:
		if err := st.RemoveTimer(name); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unknown preset kind %q", kind)
	}
	return st.Save(root)
}
