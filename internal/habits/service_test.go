package habits

import (
	"path/filepath"
	"slices"
	"testing"
)

func newTestService(t *testing.T) *Service {
	t.Helper()
	return &Service{path: filepath.Join(t.TempDir(), "habits.json"), habits: []Habit{}}
}

func TestService_ToggleCompletion(t *testing.T) {
	s := newTestService(t)

	h, err := s.Add("  Wake up at 5:30 ")
	if err != nil {
		t.Fatal(err)
	}
	if h.Name != "Wake up at 5:30" {
		t.Fatalf("name = %q, want trimmed", h.Name)
	}

	for _, date := range []string{"2026-10-07", "2026-10-05"} {
		if h, err = s.ToggleCompletion(h.ID, date); err != nil {
			t.Fatal(err)
		}
	}
	if want := []string{"2026-10-05", "2026-10-07"}; !slices.Equal(h.Completions, want) {
		t.Fatalf("completions = %v, want %v", h.Completions, want)
	}

	if h, err = s.ToggleCompletion(h.ID, "2026-10-07"); err != nil {
		t.Fatal(err)
	}
	if want := []string{"2026-10-05"}; !slices.Equal(h.Completions, want) {
		t.Fatalf("after untoggle completions = %v, want %v", h.Completions, want)
	}

	if _, err := s.ToggleCompletion(h.ID, "2026-10-07T00:00:00Z"); err != ErrInvalidDate {
		t.Fatalf("err = %v, want ErrInvalidDate", err)
	}
	if _, err := s.ToggleCompletion("missing", "2026-10-07"); err != ErrNotFound {
		t.Fatalf("err = %v, want ErrNotFound", err)
	}
}

func TestService_PersistsAcrossReload(t *testing.T) {
	s := newTestService(t)

	h, _ := s.Add("Read")
	other, _ := s.Add("Walk")
	s.ToggleCompletion(h.ID, "2026-10-07")
	if _, err := s.Rename(h.ID, "Read 20 pages"); err != nil {
		t.Fatal(err)
	}
	if err := s.Delete(other.ID); err != nil {
		t.Fatal(err)
	}

	loaded, err := load(s.path)
	if err != nil {
		t.Fatal(err)
	}
	if len(loaded) != 1 || loaded[0].Name != "Read 20 pages" || !slices.Equal(loaded[0].Completions, []string{"2026-10-07"}) {
		t.Fatalf("reloaded = %+v", loaded)
	}
}

func TestService_RejectsEmptyNames(t *testing.T) {
	s := newTestService(t)

	if _, err := s.Add("   "); err != ErrEmptyName {
		t.Fatalf("Add err = %v, want ErrEmptyName", err)
	}
	h, _ := s.Add("Stretch")
	if _, err := s.Rename(h.ID, ""); err != ErrEmptyName {
		t.Fatalf("Rename err = %v, want ErrEmptyName", err)
	}
}
