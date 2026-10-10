package calendar

import "time"

// Frequency is the unit a RecurrenceRule repeats on.
type Frequency string

const (
	FrequencyDaily   Frequency = "daily"
	FrequencyWeekly  Frequency = "weekly"
	FrequencyMonthly Frequency = "monthly"
	FrequencyYearly  Frequency = "yearly"
)

// RecurrenceRule describes how an event repeats, anchored at the event's own
// Start time. Interval is the number of Frequency units between occurrences
// (e.g. Frequency=weekly, Interval=2 means "every other week" / biweekly).
// Interval <= 0 is treated as 1.
//
// Weekdays, only meaningful for FrequencyWeekly, picks specific days of the
// week (e.g. Mon/Wed/Fri) within each repeating week instead of repeating on
// the anchor's own weekday. Weeks start on Monday, matching the frontend's
// calendar grids.
//
// Count and Until are alternative, optional end conditions; set at most one.
// If neither is set the recurrence is unbounded (expansion is still bounded
// by the date range passed to Occurrences, and by an internal safety cap).
type RecurrenceRule struct {
	Frequency Frequency      `json:"frequency"`
	Interval  int            `json:"interval"`
	Weekdays  []time.Weekday `json:"weekdays,omitempty"`
	Count     int            `json:"count,omitempty"`
	Until     *time.Time     `json:"until,omitempty"`
}

func (r RecurrenceRule) interval() int {
	if r.Interval <= 0 {
		return 1
	}
	return r.Interval
}

// maxRawOccurrences caps how many raw occurrences a single expansion will
// ever generate, so an unbounded rule can't spin forever over a huge range.
const maxRawOccurrences = 10000

// daysInMonth returns the number of days in the given month. month may be
// out of the [1,12] range; time.Date normalizes it, which is exactly what
// nextMonthlyDate below relies on.
func daysInMonth(year int, month time.Month) int {
	return time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC).Day()
}

// isLeapYear reports whether year is a leap year in the proleptic Gregorian
// calendar that time.Time uses.
func isLeapYear(year int) bool {
	return year%4 == 0 && (year%100 != 0 || year%400 == 0)
}

// nthOccurrenceDate returns the anchor-aligned date of the (n+1)th occurrence
// (n=0 is the anchor itself) for the given rule, preserving the anchor's
// time-of-day and location.
//
// Monthly/yearly recurrence clamps to the last day of the target month when
// the anchor's day doesn't exist there (e.g. "31st of every month" lands on
// Feb 28/29, not overflowing into March), matching common calendar-app
// behavior rather than Go's raw AddDate overflow semantics.
func nthOccurrenceDate(anchor time.Time, rule RecurrenceRule, n int) time.Time {
	step := n * rule.interval()
	loc := anchor.Location()
	h, m, s := anchor.Hour(), anchor.Minute(), anchor.Second()
	ns := anchor.Nanosecond()

	switch rule.Frequency {
	case FrequencyDaily:
		return anchor.AddDate(0, 0, step)
	case FrequencyWeekly:
		return anchor.AddDate(0, 0, step*7)
	case FrequencyMonthly:
		totalMonths := int(anchor.Month()) - 1 + step
		year := anchor.Year() + totalMonths/12
		month := time.Month(totalMonths%12 + 1)
		day := anchor.Day()
		if d := daysInMonth(year, month); day > d {
			day = d
		}
		return time.Date(year, month, day, h, m, s, ns, loc)
	case FrequencyYearly:
		year := anchor.Year() + step
		month := anchor.Month()
		day := anchor.Day()
		if month == time.February && day == 29 && !isLeapYear(year) {
			day = 28
		}
		return time.Date(year, month, day, h, m, s, ns, loc)
	default:
		return anchor
	}
}

// rawOccurrenceDates returns the anchor-aligned start times a recurrence
// rule produces, up to and including the first one that falls after
// rangeEnd, honoring Count/Until as early-stop conditions.
func rawOccurrenceDates(anchor time.Time, rule RecurrenceRule, rangeEnd time.Time) []time.Time {
	if rule.Frequency == FrequencyWeekly && len(rule.Weekdays) > 0 {
		return weeklyByDayDates(anchor, rule, rangeEnd)
	}

	var dates []time.Time

	for n := 0; n < maxRawOccurrences; n++ {
		if rule.Count > 0 && n >= rule.Count {
			break
		}

		date := nthOccurrenceDate(anchor, rule, n)

		if rule.Until != nil && date.After(*rule.Until) {
			break
		}

		dates = append(dates, date)

		if date.After(rangeEnd) {
			break
		}
	}

	return dates
}

// mondayIndex converts Go's Sunday-indexed Weekday (0=Sun..6=Sat) to a
// Monday-indexed offset (0=Mon..6=Sun).
func mondayIndex(d time.Weekday) int {
	return (int(d) + 6) % 7
}

// weeklyByDayDates is rawOccurrenceDates for a weekly rule with Weekdays:
// every interval() weeks, starting from the anchor's own (Monday-start)
// week, it emits each selected weekday at the anchor's time of day,
// skipping any that fall before the anchor itself. Count counts emitted
// occurrences, not weeks.
func weeklyByDayDates(anchor time.Time, rule RecurrenceRule, rangeEnd time.Time) []time.Time {
	var offsets [7]bool
	for _, d := range rule.Weekdays {
		offsets[mondayIndex(d)] = true
	}

	weekStart := anchor.AddDate(0, 0, -mondayIndex(anchor.Weekday()))

	var dates []time.Time
	for week := 0; len(dates) < maxRawOccurrences; week++ {
		monday := weekStart.AddDate(0, 0, week*7*rule.interval())

		for offset := 0; offset < 7; offset++ {
			if !offsets[offset] {
				continue
			}

			date := monday.AddDate(0, 0, offset)
			if date.Before(anchor) {
				continue
			}

			if rule.Count > 0 && len(dates) >= rule.Count {
				return dates
			}
			if rule.Until != nil && date.After(*rule.Until) {
				return dates
			}

			dates = append(dates, date)

			if date.After(rangeEnd) {
				return dates
			}
		}
	}

	return dates
}

// findException returns the Exception matching the given raw occurrence
// date (by calendar date, not exact time), if any.
func findException(exceptions []Exception, rawDate time.Time) (Exception, bool) {
	for _, ex := range exceptions {
		if sameCalendarDate(ex.OriginalDate, rawDate) {
			return ex, true
		}
	}
	return Exception{}, false
}

// isCompleted reports whether a Completion matches the given raw occurrence
// date (by calendar date, not exact time).
func isCompleted(completions []Completion, rawDate time.Time) bool {
	for _, c := range completions {
		if sameCalendarDate(c.OccurrenceDate, rawDate) {
			return true
		}
	}
	return false
}

// Occurrences expands an Event into concrete Occurrences whose (possibly
// exception-adjusted) Start falls within [rangeStart, rangeEnd], inclusive.
//
// Exceptions are resolved against the event's raw recurrence dates before
// range-filtering, so a rescheduled occurrence can move into or out of the
// requested range.
func Occurrences(e Event, rangeStart, rangeEnd time.Time) []Occurrence {
	duration := e.End.Sub(e.Start)

	var rawDates []time.Time
	if e.Recurrence == nil {
		rawDates = []time.Time{e.Start}
	} else {
		rawDates = rawOccurrenceDates(e.Start, *e.Recurrence, rangeEnd)
	}

	out := []Occurrence{} // never nil — see the note on ListOccurrences in service.go
	for _, raw := range rawDates {
		occ := Occurrence{
			Start:         raw,
			End:           raw.Add(duration),
			OriginalStart: raw,
			Done:          isCompleted(e.Completions, raw),
		}

		if ex, ok := findException(e.Exceptions, raw); ok {
			switch ex.Type {
			case ExceptionSkip:
				continue
			case ExceptionReschedule:
				if ex.NewStart == nil {
					continue
				}
				occ.Start = *ex.NewStart
				if ex.NewEnd != nil {
					occ.End = *ex.NewEnd
				} else {
					occ.End = occ.Start.Add(duration)
				}
			}
		}

		if occ.Start.Before(rangeStart) || occ.Start.After(rangeEnd) {
			continue
		}

		out = append(out, occ)
	}

	return out
}
