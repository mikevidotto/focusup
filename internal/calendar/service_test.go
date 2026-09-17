package calendar

import (
	"encoding/json"
	"path/filepath"
	"testing"
	"time"
)

// A Service with no stored events must still return `[]`, not `null`, over
// JSON — these are called directly by the frontend, which breaks (see
// CalendarWidget's onMount) if a nil Go slice marshals to `null`.
func TestService_EmptyResultsMarshalToEmptyArray(t *testing.T) {
	s := &Service{events: []Event{}}

	occurrences := s.ListOccurrences(dt(2026, 1, 1, 0, 0), dt(2026, 12, 31, 23, 59))
	if occurrences == nil {
		t.Fatal("ListOccurrences returned nil, want a non-nil empty slice")
	}
	assertMarshalsToEmptyArray(t, occurrences)

	due := s.DueReminders(time.Now())
	if due == nil {
		t.Fatal("DueReminders returned nil, want a non-nil empty slice")
	}
	assertMarshalsToEmptyArray(t, due)
}

func TestService_ToggleCompletion(t *testing.T) {
	s := &Service{
		path: filepath.Join(t.TempDir(), "calendar.json"),
		events: []Event{
			{
				ID:         "garbage-day",
				Title:      "Take out the garbage",
				Start:      dt(2026, 1, 8, 7, 0), // Thursday
				End:        dt(2026, 1, 8, 7, 15),
				Recurrence: &RecurrenceRule{Frequency: FrequencyWeekly, Interval: 1},
			},
		},
	}

	occurrenceDate := dt(2026, 1, 8, 0, 0)

	updated, err := s.ToggleCompletion("garbage-day", occurrenceDate)
	if err != nil {
		t.Fatalf("first toggle: %v", err)
	}
	if len(updated.Completions) != 1 {
		t.Fatalf("after first toggle: got %d completions, want 1", len(updated.Completions))
	}

	occs := s.ListOccurrences(dt(2026, 1, 8, 0, 0), dt(2026, 1, 8, 23, 59))
	if len(occs) != 1 || !occs[0].Done {
		t.Fatalf("ListOccurrences after toggle: got %+v, want exactly one Done occurrence", occs)
	}

	updated, err = s.ToggleCompletion("garbage-day", occurrenceDate)
	if err != nil {
		t.Fatalf("second toggle: %v", err)
	}
	if len(updated.Completions) != 0 {
		t.Fatalf("after second toggle (un-complete): got %d completions, want 0", len(updated.Completions))
	}

	if _, err := s.ToggleCompletion("does-not-exist", occurrenceDate); err != ErrNotFound {
		t.Errorf("toggling an unknown event: got err %v, want ErrNotFound", err)
	}
}

func assertMarshalsToEmptyArray(t *testing.T, v interface{}) {
	t.Helper()

	data, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}

	if string(data) != "[]" {
		t.Errorf("marshaled to %q, want \"[]\"", data)
	}
}
