package journal

import (
	"path/filepath"
	"testing"
)

func newTestService(t *testing.T) *Service {
	t.Helper()
	return &Service{path: filepath.Join(t.TempDir(), "journal.json"), entries: []Entry{}}
}

func prompts(answers ...string) []PromptAnswer {
	out := make([]PromptAnswer, len(answers))
	for i, a := range answers {
		out[i] = PromptAnswer{Prompt: "Prompt", Answer: a}
	}
	return out
}

func TestService_SaveUpserts(t *testing.T) {
	s := newTestService(t)

	if _, err := s.Save("2026-10-08", prompts("first", ""), "  body \n\n"); err != nil {
		t.Fatal(err)
	}
	e, err := s.Save("2026-10-08", prompts("second", "more"), "")
	if err != nil {
		t.Fatal(err)
	}

	if got := s.List(); len(got) != 1 {
		t.Fatalf("len = %d, want 1 entry after upsert", len(got))
	}
	if e.Prompts[0].Answer != "second" || e.Prompts[1].Answer != "more" || e.Body != "" {
		t.Fatalf("entry = %+v", e)
	}
}

func TestService_TrimsTrailingWhitespace(t *testing.T) {
	s := newTestService(t)

	e, _ := s.Save("2026-10-08", prompts("  indented\n"), "  body \n\n")
	if e.Body != "  body" || e.Prompts[0].Answer != "  indented" {
		t.Fatalf("entry = %+v, want trailing whitespace trimmed only", e)
	}
}

func TestService_BlankSaveRemovesEntry(t *testing.T) {
	s := newTestService(t)

	s.Save("2026-10-08", prompts("something"), "")
	if _, err := s.Save("2026-10-08", prompts("   ", ""), "\n"); err != nil {
		t.Fatal(err)
	}
	if got := s.List(); len(got) != 0 {
		t.Fatalf("entries = %+v, want blank save to remove the entry", got)
	}
	if _, err := s.Get("2026-10-08"); err != ErrNotFound {
		t.Fatalf("err = %v, want ErrNotFound", err)
	}

	// A blank save for a day that never had an entry is a no-op, not an error.
	if _, err := s.Save("2026-10-09", nil, ""); err != nil {
		t.Fatal(err)
	}
}

func TestService_RejectsInvalidDate(t *testing.T) {
	s := newTestService(t)

	if _, err := s.Save("2026-10-08T00:00:00Z", nil, "x"); err != ErrInvalidDate {
		t.Fatalf("err = %v, want ErrInvalidDate", err)
	}
	if err := s.Delete("2026-10-08"); err != ErrNotFound {
		t.Fatalf("err = %v, want ErrNotFound", err)
	}
}

func TestService_ListNewestFirstAndPersists(t *testing.T) {
	s := newTestService(t)

	for _, date := range []string{"2026-10-06", "2026-10-08", "2026-10-07", "2026-09-30"} {
		if _, err := s.Save(date, nil, "entry "+date); err != nil {
			t.Fatal(err)
		}
	}
	if err := s.Delete("2026-10-07"); err != nil {
		t.Fatal(err)
	}

	want := []string{"2026-10-08", "2026-10-06", "2026-09-30"}
	got := s.List()
	for i, e := range got {
		if e.Date != want[i] {
			t.Fatalf("order = %v, want %v", got, want)
		}
	}

	loaded, err := load(s.path)
	if err != nil {
		t.Fatal(err)
	}
	if len(loaded) != 3 || loaded[0].Date != "2026-10-08" || loaded[0].Body != "entry 2026-10-08" {
		t.Fatalf("reloaded = %+v", loaded)
	}
}
