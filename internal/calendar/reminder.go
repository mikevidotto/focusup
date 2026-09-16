package calendar

import "time"

// Reminder fires a notification LeadTime before an occurrence's Start
// (e.g. LeadTime = 24*time.Hour for "notify 1 day before").
type Reminder struct {
	ID       string        `json:"id"`
	LeadTime time.Duration `json:"leadTime"`
}

// ReminderFireTime returns the moment a Reminder should fire for an
// occurrence starting at occurrenceStart.
func ReminderFireTime(occurrenceStart time.Time, r Reminder) time.Time {
	return occurrenceStart.Add(-r.LeadTime)
}

// IsReminderDue reports whether a Reminder for the given occurrence is due
// at `now`: its fire time has passed but the occurrence itself hasn't
// started yet. Once the occurrence starts, its reminder is no longer due.
func IsReminderDue(occurrenceStart time.Time, r Reminder, now time.Time) bool {
	fireTime := ReminderFireTime(occurrenceStart, r)
	return !now.Before(fireTime) && now.Before(occurrenceStart)
}
