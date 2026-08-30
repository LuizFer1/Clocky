package pomodoro

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/luisf/clocky/internal/clockface"
	"github.com/luisf/clocky/internal/notify"
)

// Config holds pomodoro timing and behavior.
type Config struct {
	Focus  time.Duration
	Break  time.Duration
	Long   time.Duration
	Cycles int
	Auto   bool
}

// Phase is one focus or break segment in a planned run.
type Phase struct {
	Name     string // "FOCUS" | "BREAK" | "LONG BREAK"
	Duration time.Duration
	Cycle    int // 1-based focus index for FOCUS; completed focus count for breaks
}

// PlanPhases builds the phase sequence for one set of cycles ending in a long break.
func PlanPhases(cfg Config) []Phase {
	if cfg.Cycles < 1 {
		return nil
	}
	phases := make([]Phase, 0, cfg.Cycles*2)
	for i := 1; i <= cfg.Cycles; i++ {
		phases = append(phases, Phase{Name: "FOCUS", Duration: cfg.Focus, Cycle: i})
		if i < cfg.Cycles {
			phases = append(phases, Phase{Name: "BREAK", Duration: cfg.Break, Cycle: i})
		} else {
			phases = append(phases, Phase{Name: "LONG BREAK", Duration: cfg.Long, Cycle: i})
		}
	}
	return phases
}

// Hooks injects I/O and timing for Run (and tests).
type Hooks struct {
	Notify notify.Notifier
	Sleep  func(d time.Duration)
	Now    func() time.Time
	In     io.Reader
	Out    io.Writer
	Width  int
}

func (h Hooks) withDefaults() Hooks {
	if h.Sleep == nil {
		h.Sleep = time.Sleep
	}
	if h.Now == nil {
		h.Now = time.Now
	}
	if h.Out == nil {
		h.Out = os.Stdout
	}
	if h.Notify == nil {
		h.Notify = notify.Default{}
	}
	if h.Width == 0 {
		h.Width = 21
	}
	if h.In == nil {
		h.In = os.Stdin
	}
	return h
}

// Run executes the pomodoro phase plan using hooks for timing and I/O.
func Run(cfg Config, h Hooks) error {
	h = h.withDefaults()
	phases := PlanPhases(cfg)
	for i, phase := range phases {
		if err := runPhase(cfg, phase, h); err != nil {
			return err
		}
		title, body := phaseEndMessage(phase)
		if err := notify.All(h.Notify, title, body); err != nil {
			return err
		}
		if !cfg.Auto && i < len(phases)-1 {
			if err := waitEnter(h.In, h.Out); err != nil {
				return err
			}
		}
	}
	return nil
}

func runPhase(cfg Config, phase Phase, h Hooks) error {
	remaining := phase.Duration
	total := phase.Duration
	if total <= 0 {
		total = time.Second
	}
	for remaining > 0 {
		if _, err := fmt.Fprint(h.Out, "\033[H\033[2J"); err != nil {
			return err
		}
		if _, err := fmt.Fprint(h.Out, clockface.Render(remaining, total, h.Width)); err != nil {
			return err
		}
		label := clockface.RenderCompact(phase.Name, phase.Cycle, cfg.Cycles, remaining)
		if _, err := fmt.Fprintln(h.Out, label); err != nil {
			return err
		}
		step := time.Second
		if remaining < step {
			step = remaining
		}
		h.Sleep(step)
		remaining -= step
	}
	return nil
}

func phaseEndMessage(phase Phase) (title, body string) {
	switch phase.Name {
	case "FOCUS":
		return "Focus complete", fmt.Sprintf("Focus session %d finished", phase.Cycle)
	case "BREAK":
		return "Break complete", fmt.Sprintf("Short break after focus %d finished", phase.Cycle)
	default:
		return "Long break complete", fmt.Sprintf("Long break after focus %d finished", phase.Cycle)
	}
}

func waitEnter(in io.Reader, out io.Writer) error {
	if _, err := fmt.Fprint(out, "Press Enter to continue..."); err != nil {
		return err
	}
	_, err := bufio.NewReader(in).ReadString('\n')
	return err
}
