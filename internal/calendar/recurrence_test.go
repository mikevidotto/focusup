package calendar

import (
	"testing"
	"time"
)

func dt(y int, m time.Month, d, h, min int) time.Time {
	return time.Date(y, m, d, h, min, 0, 0, time.UTC)
}

func occurrenceStarts(occs []Occurrence) []time.Time {
	starts := make([]time.Time, len(occs))
	for i, o := range occs {
		starts[i] = o.Start
	}
	return starts
}

func assertTimes(t *testing.T, got, want []time.Time) {
	t.Helper()

	if len(got) != len(want) {
		t.Fatalf("got %d occurrences %v, want %d %v", len(got), got, len(want), want)
	}

	for i := range want {
		if !got[i].Equal(want[i]) {
			t.Errorf("occurrence %d: got %v, want %v", i, got[i], want[i])
		}
	}
}

func TestOccurrences_Recurrence(t *testing.T) {
	tests := []struct {
		name       string
		event      Event
		rangeStart time.Time
		rangeEnd   time.Time
		want       []time.Time
	}{
		{
			name: "weekly",
			event: Event{
				Start:      dt(2026, 1, 5, 9, 0), // Monday
				End:        dt(2026, 1, 5, 10, 0),
				Recurrence: &RecurrenceRule{Frequency: FrequencyWeekly, Interval: 1},
			},
			rangeStart: dt(2026, 1, 1, 0, 0),
			rangeEnd:   dt(2026, 1, 31, 23, 59),
			want: []time.Time{
				dt(2026, 1, 5, 9, 0),
				dt(2026, 1, 12, 9, 0),
				dt(2026, 1, 19, 9, 0),
				dt(2026, 1, 26, 9, 0),
			},
		},
		{
			name: "biweekly garbage day",
			event: Event{
				Start:      dt(2026, 1, 1, 7, 0), // Thursday
				End:        dt(2026, 1, 1, 7, 30),
				Recurrence: &RecurrenceRule{Frequency: FrequencyWeekly, Interval: 2},
			},
			rangeStart: dt(2026, 1, 1, 0, 0),
			rangeEnd:   dt(2026, 2, 28, 23, 59),
			want: []time.Time{
				dt(2026, 1, 1, 7, 0),
				dt(2026, 1, 15, 7, 0),
				dt(2026, 1, 29, 7, 0),
				dt(2026, 2, 12, 7, 0),
				dt(2026, 2, 26, 7, 0),
			},
		},
		{
			name: "monthly on the 15th",
			event: Event{
				Start:      dt(2026, 1, 15, 12, 0),
				End:        dt(2026, 1, 15, 13, 0),
				Recurrence: &RecurrenceRule{Frequency: FrequencyMonthly, Interval: 1},
			},
			rangeStart: dt(2026, 1, 1, 0, 0),
			rangeEnd:   dt(2026, 3, 31, 23, 59),
			want: []time.Time{
				dt(2026, 1, 15, 12, 0),
				dt(2026, 2, 15, 12, 0),
				dt(2026, 3, 15, 12, 0),
			},
		},
		{
			name: "monthly month-end clamps instead of overflowing",
			event: Event{
				Start:      dt(2026, 1, 31, 8, 0),
				End:        dt(2026, 1, 31, 9, 0),
				Recurrence: &RecurrenceRule{Frequency: FrequencyMonthly, Interval: 1},
			},
			rangeStart: dt(2026, 1, 1, 0, 0),
			rangeEnd:   dt(2026, 4, 30, 23, 59),
			want: []time.Time{
				dt(2026, 1, 31, 8, 0),
				dt(2026, 2, 28, 8, 0), // 2026 is not a leap year
				dt(2026, 3, 31, 8, 0),
				dt(2026, 4, 30, 8, 0),
			},
		},
		{
			name: "monthly month-end in a leap year",
			event: Event{
				Start:      dt(2028, 1, 31, 8, 0),
				End:        dt(2028, 1, 31, 9, 0),
				Recurrence: &RecurrenceRule{Frequency: FrequencyMonthly, Interval: 1},
			},
			rangeStart: dt(2028, 1, 1, 0, 0),
			rangeEnd:   dt(2028, 2, 29, 23, 59),
			want: []time.Time{
				dt(2028, 1, 31, 8, 0),
				dt(2028, 2, 29, 8, 0), // 2028 is a leap year
			},
		},
		{
			name: "count limits total occurrences",
			event: Event{
				Start:      dt(2026, 1, 5, 9, 0),
				End:        dt(2026, 1, 5, 10, 0),
				Recurrence: &RecurrenceRule{Frequency: FrequencyWeekly, Interval: 1, Count: 2},
			},
			rangeStart: dt(2026, 1, 1, 0, 0),
			rangeEnd:   dt(2026, 12, 31, 23, 59),
			want: []time.Time{
				dt(2026, 1, 5, 9, 0),
				dt(2026, 1, 12, 9, 0),
			},
		},
		{
			name: "exception skips a single occurrence",
			event: Event{
				Start:      dt(2026, 1, 5, 9, 0),
				End:        dt(2026, 1, 5, 10, 0),
				Recurrence: &RecurrenceRule{Frequency: FrequencyWeekly, Interval: 1},
				Exceptions: []Exception{
					{OriginalDate: dt(2026, 1, 12, 0, 0), Type: ExceptionSkip},
				},
			},
			rangeStart: dt(2026, 1, 1, 0, 0),
			rangeEnd:   dt(2026, 1, 31, 23, 59),
			want: []time.Time{
				dt(2026, 1, 5, 9, 0),
				dt(2026, 1, 19, 9, 0),
				dt(2026, 1, 26, 9, 0),
			},
		},
		{
			name: "exception reschedules a single occurrence within range",
			event: Event{
				Start:      dt(2026, 1, 5, 9, 0),
				End:        dt(2026, 1, 5, 10, 0),
				Recurrence: &RecurrenceRule{Frequency: FrequencyWeekly, Interval: 1},
				Exceptions: []Exception{
					{
						OriginalDate: dt(2026, 1, 12, 0, 0),
						Type:         ExceptionReschedule,
						NewStart:     ptr(dt(2026, 1, 14, 15, 0)),
					},
				},
			},
			rangeStart: dt(2026, 1, 1, 0, 0),
			rangeEnd:   dt(2026, 1, 31, 23, 59),
			want: []time.Time{
				dt(2026, 1, 5, 9, 0),
				dt(2026, 1, 14, 15, 0), // moved from Jan 12 to Jan 14
				dt(2026, 1, 19, 9, 0),
				dt(2026, 1, 26, 9, 0),
			},
		},
		{
			name: "exception reschedules an occurrence into range from outside it",
			event: Event{
				Start:      dt(2025, 12, 29, 9, 0), // Monday, one week before range
				End:        dt(2025, 12, 29, 10, 0),
				Recurrence: &RecurrenceRule{Frequency: FrequencyWeekly, Interval: 1},
				Exceptions: []Exception{
					{
						OriginalDate: dt(2025, 12, 29, 0, 0),
						Type:         ExceptionReschedule,
						NewStart:     ptr(dt(2026, 1, 2, 9, 0)),
					},
				},
			},
			rangeStart: dt(2026, 1, 1, 0, 0),
			rangeEnd:   dt(2026, 1, 5, 23, 59),
			want: []time.Time{
				dt(2026, 1, 2, 9, 0),
				dt(2026, 1, 5, 9, 0),
			},
		},
		{
			name: "non-recurring event within range",
			event: Event{
				Start: dt(2026, 3, 3, 9, 0),
				End:   dt(2026, 3, 3, 10, 0),
			},
			rangeStart: dt(2026, 1, 1, 0, 0),
			rangeEnd:   dt(2026, 12, 31, 23, 59),
			want:       []time.Time{dt(2026, 3, 3, 9, 0)},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := Occurrences(tc.event, tc.rangeStart, tc.rangeEnd)
			assertTimes(t, occurrenceStarts(got), tc.want)
		})
	}
}

// Occurrences is JSON-marshaled straight to the frontend, where a nil slice
// becomes `null` instead of `[]` and breaks any array method called on it
// (this broke the dashboard widget in practice — see service.go's
// ListOccurrences). Guard against regressing back to a nil result.
func TestOccurrences_NeverReturnsNilWhenEmpty(t *testing.T) {
	e := Event{
		Start: dt(2026, 1, 5, 9, 0),
		End:   dt(2026, 1, 5, 10, 0),
	}

	got := Occurrences(e, dt(2030, 1, 1, 0, 0), dt(2030, 1, 2, 0, 0))

	if got == nil {
		t.Fatal("Occurrences returned nil, want a non-nil empty slice")
	}
	if len(got) != 0 {
		t.Fatalf("got %d occurrences, want 0", len(got))
	}
}

func TestOccurrences_Completions(t *testing.T) {
	e := Event{
		Start:      dt(2026, 1, 5, 9, 0), // Monday
		End:        dt(2026, 1, 5, 9, 30),
		Recurrence: &RecurrenceRule{Frequency: FrequencyWeekly, Interval: 1},
		Completions: []Completion{
			{OccurrenceDate: dt(2026, 1, 12, 0, 0), CompletedAt: dt(2026, 1, 12, 9, 15)},
		},
	}

	occs := Occurrences(e, dt(2026, 1, 1, 0, 0), dt(2026, 1, 19, 23, 59))
	if len(occs) != 3 {
		t.Fatalf("got %d occurrences, want 3", len(occs))
	}

	want := map[string]bool{
		occs[0].Start.String(): false, // Jan 5
		occs[1].Start.String(): true,  // Jan 12 — completed
		occs[2].Start.String(): false, // Jan 19
	}

	for _, occ := range occs {
		if occ.Done != want[occ.Start.String()] {
			t.Errorf("occurrence %v: Done = %v, want %v", occ.Start, occ.Done, want[occ.Start.String()])
		}
	}
}

// A Completion is matched against an occurrence's raw (pre-exception) date,
// so it still applies after that occurrence gets rescheduled — the same
// stable identity Exception itself is matched against.
func TestOccurrences_CompletionSurvivesReschedule(t *testing.T) {
	e := Event{
		Start:      dt(2026, 1, 5, 9, 0),
		End:        dt(2026, 1, 5, 9, 30),
		Recurrence: &RecurrenceRule{Frequency: FrequencyWeekly, Interval: 1},
		Exceptions: []Exception{
			{
				OriginalDate: dt(2026, 1, 12, 0, 0),
				Type:         ExceptionReschedule,
				NewStart:     ptr(dt(2026, 1, 14, 15, 0)),
			},
		},
		Completions: []Completion{
			{OccurrenceDate: dt(2026, 1, 12, 0, 0), CompletedAt: dt(2026, 1, 12, 9, 15)},
		},
	}

	occs := Occurrences(e, dt(2026, 1, 1, 0, 0), dt(2026, 1, 19, 23, 59))

	var rescheduled *Occurrence
	for i := range occs {
		if occs[i].Start.Equal(dt(2026, 1, 14, 15, 0)) {
			rescheduled = &occs[i]
		}
	}

	if rescheduled == nil {
		t.Fatalf("did not find the rescheduled occurrence in %v", occurrenceStarts(occs))
	}
	if !rescheduled.Done {
		t.Error("rescheduled occurrence lost its completion, want Done = true")
	}
}

func TestOccurrences_RescheduleKeepsDurationWhenNewEndOmitted(t *testing.T) {
	e := Event{
		Start:      dt(2026, 1, 5, 9, 0),
		End:        dt(2026, 1, 5, 9, 30), // 30 minute event
		Recurrence: &RecurrenceRule{Frequency: FrequencyWeekly, Interval: 1},
		Exceptions: []Exception{
			{
				OriginalDate: dt(2026, 1, 12, 0, 0),
				Type:         ExceptionReschedule,
				NewStart:     ptr(dt(2026, 1, 13, 14, 0)),
			},
		},
	}

	occs := Occurrences(e, dt(2026, 1, 12, 0, 0), dt(2026, 1, 13, 23, 59))
	if len(occs) != 1 {
		t.Fatalf("got %d occurrences, want 1", len(occs))
	}

	wantEnd := dt(2026, 1, 13, 14, 30)
	if !occs[0].End.Equal(wantEnd) {
		t.Errorf("rescheduled end = %v, want %v", occs[0].End, wantEnd)
	}
}

func ptr(t time.Time) *time.Time {
	return &t
}
