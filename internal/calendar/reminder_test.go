package calendar

import (
	"testing"
	"time"
)

func TestIsReminderDue(t *testing.T) {
	occurrenceStart := dt(2026, 1, 10, 9, 0)
	reminder := Reminder{LeadTime: 24 * time.Hour} // notify 1 day before

	tests := []struct {
		name string
		now  time.Time
		want bool
	}{
		{
			name: "well before the fire time",
			now:  dt(2026, 1, 8, 9, 0),
			want: false,
		},
		{
			name: "exactly at the fire time",
			now:  dt(2026, 1, 9, 9, 0),
			want: true,
		},
		{
			name: "between fire time and occurrence start",
			now:  dt(2026, 1, 9, 20, 0),
			want: true,
		},
		{
			name: "exactly at occurrence start",
			now:  dt(2026, 1, 10, 9, 0),
			want: false,
		},
		{
			name: "after the occurrence has started",
			now:  dt(2026, 1, 10, 10, 0),
			want: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := IsReminderDue(occurrenceStart, reminder, tc.now)
			if got != tc.want {
				t.Errorf("IsReminderDue(%v) = %v, want %v", tc.now, got, tc.want)
			}
		})
	}
}

func TestReminderFireTime(t *testing.T) {
	occurrenceStart := dt(2026, 1, 10, 9, 0)

	tests := []struct {
		name     string
		leadTime time.Duration
		want     time.Time
	}{
		{name: "1 day before", leadTime: 24 * time.Hour, want: dt(2026, 1, 9, 9, 0)},
		{name: "1 hour before", leadTime: time.Hour, want: dt(2026, 1, 10, 8, 0)},
		{name: "0 lead time fires at start", leadTime: 0, want: dt(2026, 1, 10, 9, 0)},
		{name: "1 week before", leadTime: 7 * 24 * time.Hour, want: dt(2026, 1, 3, 9, 0)},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := ReminderFireTime(occurrenceStart, Reminder{LeadTime: tc.leadTime})
			if !got.Equal(tc.want) {
				t.Errorf("ReminderFireTime = %v, want %v", got, tc.want)
			}
		})
	}
}
