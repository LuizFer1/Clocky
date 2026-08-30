package state

import (
	"os"
	"path/filepath"
)

// EnsureDir creates root and any missing parents.
func EnsureDir(root string) error {
	return os.MkdirAll(root, 0o755)
}

// Path joins root with a state file name.
func Path(root, name string) string {
	return filepath.Join(root, name)
}

// DefaultRoot returns ~/.clocky.
func DefaultRoot() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".clocky"), nil
}
