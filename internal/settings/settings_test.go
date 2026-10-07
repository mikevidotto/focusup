package settings

import (
	"path/filepath"
	"testing"
)

func TestService_SetThemePersists(t *testing.T) {
	path := filepath.Join(t.TempDir(), "settings.json")

	loaded, err := load(path)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Theme != "dark" {
		t.Fatalf("default theme = %q, want dark", loaded.Theme)
	}

	s := &Service{path: path, settings: loaded}
	if _, err := s.SetTheme("light"); err != nil {
		t.Fatal(err)
	}
	if _, err := s.SetTheme("neon"); err != ErrInvalidTheme {
		t.Fatalf("err = %v, want ErrInvalidTheme", err)
	}

	reloaded, err := load(path)
	if err != nil {
		t.Fatal(err)
	}
	if reloaded.Theme != "light" {
		t.Fatalf("reloaded theme = %q, want light", reloaded.Theme)
	}
}
