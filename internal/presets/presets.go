package presets

import (
	"errors"
	"fmt"
	"os"

	"github.com/LuizFer1/Clocky/internal/state"
)

const fileName = "presets.json"

type PomodoroPreset struct {
	Name   string `json:"name"`
	Focus  int    `json:"focus"` // minutes
	Break  int    `json:"break"`
	Long   int    `json:"long"`
	Cycles int    `json:"cycles"`
	Auto   bool   `json:"auto"`
}

type TimerPreset struct {
	Name    string `json:"name"`
	Seconds int64  `json:"seconds"`
}

type Store struct {
	Pomodoros []PomodoroPreset `json:"pomodoros"`
	Timers    []TimerPreset    `json:"timers"`
}

// Load reads presets.json under root. Missing file yields an empty store.
func Load(root string) (*Store, error) {
	path := state.Path(root, fileName)
	var s Store
	err := state.ReadJSON(path, &s)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return &Store{}, nil
		}
		return nil, err
	}
	if s.Pomodoros == nil {
		s.Pomodoros = []PomodoroPreset{}
	}
	if s.Timers == nil {
		s.Timers = []TimerPreset{}
	}
	return &s, nil
}

// Save writes the store to presets.json under root.
func (s *Store) Save(root string) error {
	if err := state.EnsureDir(root); err != nil {
		return err
	}
	return state.WriteJSON(state.Path(root, fileName), s)
}

func (s *Store) UpsertPomodoro(p PomodoroPreset) {
	for i := range s.Pomodoros {
		if s.Pomodoros[i].Name == p.Name {
			s.Pomodoros[i] = p
			return
		}
	}
	s.Pomodoros = append(s.Pomodoros, p)
}

func (s *Store) UpsertTimer(t TimerPreset) {
	for i := range s.Timers {
		if s.Timers[i].Name == t.Name {
			s.Timers[i] = t
			return
		}
	}
	s.Timers = append(s.Timers, t)
}

func (s *Store) GetPomodoro(name string) (PomodoroPreset, bool) {
	for _, p := range s.Pomodoros {
		if p.Name == name {
			return p, true
		}
	}
	return PomodoroPreset{}, false
}

func (s *Store) GetTimer(name string) (TimerPreset, bool) {
	for _, t := range s.Timers {
		if t.Name == name {
			return t, true
		}
	}
	return TimerPreset{}, false
}

func (s *Store) RemovePomodoro(name string) error {
	for i, p := range s.Pomodoros {
		if p.Name == name {
			s.Pomodoros = append(s.Pomodoros[:i], s.Pomodoros[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("pomodoro preset %q not found", name)
}

func (s *Store) RemoveTimer(name string) error {
	for i, t := range s.Timers {
		if t.Name == name {
			s.Timers = append(s.Timers[:i], s.Timers[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("timer preset %q not found", name)
}
