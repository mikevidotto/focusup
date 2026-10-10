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

func newTestService(t *testing.T) *Service {
	t.Helper()
	return &Service{path: filepath.Join(t.TempDir(), "calendar.json"), events: []Event{}}
}

func TestService_AddValidates(t *testing.T) {
	s := newTestService(t)
	start := dt(2026, 1, 8, 9, 0)
	until := dt(2026, 2, 1, 0, 0)

	bad := map[string]EventInput{
		"empty title":         {Title: "  ", Start: start, End: start},
		"end before start":    {Title: "x", Start: start, End: start.Add(-time.Hour)},
		"negative reminder":   {Title: "x", Start: start, End: start, ReminderLeadSeconds: []int{-60}},
		"unknown frequency":   {Title: "x", Start: start, End: start, Recurrence: &RecurrenceRule{Frequency: "hourly"}},
		"count and until":     {Title: "x", Start: start, End: start, Recurrence: &RecurrenceRule{Frequency: FrequencyDaily, Count: 3, Until: &until}},
		"weekdays non-weekly": {Title: "x", Start: start, End: start, Recurrence: &RecurrenceRule{Frequency: FrequencyDaily, Weekdays: []time.Weekday{time.Monday}}},
	}
	for name, in := range bad {
		if _, err := s.Add(in); err == nil {
			t.Errorf("%s: Add succeeded, want an error", name)
		}
	}
	if len(s.events) != 0 {
		t.Fatalf("invalid adds stored %d events", len(s.events))
	}

	e, err := s.Add(EventInput{
		Title: " Gym ", Start: start, End: start.Add(time.Hour),
		Recurrence:          &RecurrenceRule{Frequency: FrequencyWeekly},
		ReminderLeadSeconds: []int{600},
	})
	if err != nil {
		t.Fatalf("valid Add: %v", err)
	}
	if e.Title != "Gym" || e.Recurrence.Interval != 1 || len(e.Reminders) != 1 || e.Reminders[0].LeadTime != 10*time.Minute {
		t.Errorf("Add stored %+v", e)
	}
	if e.Start.Location() != time.Local {
		t.Errorf("Start location = %v, want Local", e.Start.Location())
	}
}

func TestService_UpdateKeepsHistoryAndReminderIDs(t *testing.T) {
	s := newTestService(t)
	start := dt(2026, 1, 8, 9, 0)

	e, err := s.Add(EventInput{
		Title: "Standup", Start: start, End: start,
		Recurrence:          &RecurrenceRule{Frequency: FrequencyDaily},
		ReminderLeadSeconds: []int{600, 3600},
	})
	if err != nil {
		t.Fatal(err)
	}
	keptID := e.Reminders[0].ID

	if _, err := s.ToggleCompletion(e.ID, start); err != nil {
		t.Fatal(err)
	}
	if _, err := s.SkipOccurrence(e.ID, start.AddDate(0, 0, 1)); err != nil {
		t.Fatal(err)
	}

	updated, err := s.Update(e.ID, EventInput{
		Title: "Daily standup", Location: "Zoom", Important: true,
		Start: start, End: start.Add(15 * time.Minute),
		Recurrence:          &RecurrenceRule{Frequency: FrequencyDaily},
		ReminderLeadSeconds: []int{600, 0},
	})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}

	if updated.ID != e.ID || !updated.CreatedAt.Equal(e.CreatedAt) {
		t.Errorf("Update changed identity: %+v", updated)
	}
	if updated.Title != "Daily standup" || updated.Location != "Zoom" || !updated.Important {
		t.Errorf("Update didn't apply fields: %+v", updated)
	}
	if len(updated.Completions) != 1 || len(updated.Exceptions) != 1 {
		t.Errorf("Update dropped history: completions=%d exceptions=%d", len(updated.Completions), len(updated.Exceptions))
	}
	if len(updated.Reminders) != 2 || updated.Reminders[0].ID != keptID {
		t.Errorf("unchanged reminder lost its ID: %+v (want first ID %s)", updated.Reminders, keptID)
	}

	if _, err := s.Update("does-not-exist", EventInput{Title: "x", Start: start, End: start}); err != ErrNotFound {
		t.Errorf("updating an unknown event: got err %v, want ErrNotFound", err)
	}
}

func TestService_SkipOccurrence(t *testing.T) {
	s := newTestService(t)
	start := dt(2026, 1, 5, 21, 0)

	e, err := s.Add(EventInput{
		Title: "Trash", Start: start, End: start,
		Recurrence: &RecurrenceRule{Frequency: FrequencyWeekly},
	})
	if err != nil {
		t.Fatal(err)
	}

	occs := s.ListOccurrences(dt(2026, 1, 1, 0, 0), dt(2026, 1, 31, 0, 0))
	if len(occs) != 4 || !occs[0].Recurring {
		t.Fatalf("before skip: got %+v", occs)
	}

	if _, err := s.SkipOccurrence(e.ID, occs[1].OriginalStart); err != nil {
		t.Fatal(err)
	}

	after := s.ListOccurrences(dt(2026, 1, 1, 0, 0), dt(2026, 1, 31, 0, 0))
	assertTimes(t, []time.Time{after[0].Start, after[1].Start, after[2].Start},
		[]time.Time{occs[0].Start, occs[2].Start, occs[3].Start})
	if len(after) != 3 {
		t.Errorf("after skip: got %d occurrences, want 3", len(after))
	}
}
