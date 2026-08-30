package tui

import (
	"fmt"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/LuizFer1/Clocky/internal/state"
)

// Run starts the interactive Terminal UI.
func Run() error {
	root, err := state.DefaultRoot()
	if err != nil {
		return err
	}
	if err := state.EnsureDir(root); err != nil {
		return err
	}
	m := initialModel(Dependencies{
		Root: root,
		Now:  time.Now,
		Exe:  os.Executable,
	})
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("tui: %w", err)
	}
	return nil
}
