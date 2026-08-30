package tui

import (
	"fmt"
	"time"

	"github.com/LuizFer1/Clocky/internal/duration"
	"github.com/LuizFer1/Clocky/internal/presets"
)

const (
	kindPomodoro = "pomodoro"
	kindTimer    = "timer"
)

type presetItem struct {
	Kind    string
	Name    string
	Summary string
	Pomo    *presets.PomodoroPreset
	Tim     *presets.TimerPreset
}

func buildPresetItems(s *presets.Store) []presetItem {
	if s == nil {
		return nil
	}
	items := make([]presetItem, 0, len(s.Pomodoros)+len(s.Timers))
	for i := range s.Pomodoros {
		p := s.Pomodoros[i]
		pCopy := p
		auto := ""
		if p.Auto {
			auto = " auto"
		}
		items = append(items, presetItem{
			Kind:    kindPomodoro,
			Name:    p.Name,
			Summary: fmt.Sprintf("pomo  %d/%d/%d x%d%s", p.Focus, p.Break, p.Long, p.Cycles, auto),
			Pomo:    &pCopy,
		})
	}
	for i := range s.Timers {
		t := s.Timers[i]
		tCopy := t
		items = append(items, presetItem{
			Kind:    kindTimer,
			Name:    t.Name,
			Summary: fmt.Sprintf("timer %s", duration.Format(time.Duration(t.Seconds)*time.Second)),
			Tim:     &tCopy,
		})
	}
	return items
}

func loadPresetItems(root string) ([]presetItem, error) {
	st, err := presets.Load(root)
	if err != nil {
		return nil, err
	}
	return buildPresetItems(st), nil
}
