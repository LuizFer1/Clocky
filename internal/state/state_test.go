package state

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestEnsureDir(t *testing.T) {
	root := filepath.Join(t.TempDir(), "nested", ".clocky")
	if err := EnsureDir(root); err != nil {
		t.Fatalf("EnsureDir: %v", err)
	}
	info, err := os.Stat(root)
	if err != nil {
		t.Fatalf("Stat: %v", err)
	}
	if !info.IsDir() {
		t.Fatalf("expected directory at %s", root)
	}
	// Idempotent.
	if err := EnsureDir(root); err != nil {
		t.Fatalf("EnsureDir second call: %v", err)
	}
}

func TestWriteJSONReadJSONRoundTrip(t *testing.T) {
	root := t.TempDir()
	if err := EnsureDir(root); err != nil {
		t.Fatalf("EnsureDir: %v", err)
	}
	path := Path(root, "sample.json")

	type sample struct {
		Name  string `json:"name"`
		Count int    `json:"count"`
	}
	want := sample{Name: "focus", Count: 4}
	if err := WriteJSON(path, want); err != nil {
		t.Fatalf("WriteJSON: %v", err)
	}

	var got sample
	if err := ReadJSON(path, &got); err != nil {
		t.Fatalf("ReadJSON: %v", err)
	}
	if got != want {
		t.Fatalf("got %+v want %+v", got, want)
	}

	// Overwrite existing file (critical on Windows where rename cannot replace).
	want.Count = 9
	if err := WriteJSON(path, want); err != nil {
		t.Fatalf("WriteJSON overwrite: %v", err)
	}
	if err := ReadJSON(path, &got); err != nil {
		t.Fatalf("ReadJSON after overwrite: %v", err)
	}
	if got != want {
		t.Fatalf("after overwrite got %+v want %+v", got, want)
	}
}

func TestReadJSONMissing(t *testing.T) {
	path := Path(t.TempDir(), "missing.json")
	var v struct{}
	err := ReadJSON(path, &v)
	if !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("got %v, want errors.Is(..., os.ErrNotExist)", err)
	}
}

func TestPath(t *testing.T) {
	got := Path("/tmp/.clocky", "presets.json")
	want := filepath.Join("/tmp/.clocky", "presets.json")
	if got != want {
		t.Fatalf("Path = %q want %q", got, want)
	}
}

func TestDefaultRoot(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skipf("UserHomeDir: %v", err)
	}
	got, err := DefaultRoot()
	if err != nil {
		t.Fatalf("DefaultRoot: %v", err)
	}
	want := filepath.Join(home, ".clocky")
	if got != want {
		t.Fatalf("DefaultRoot = %q want %q", got, want)
	}
}
