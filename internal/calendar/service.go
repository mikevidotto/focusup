package calendar

import (
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
)

type Service struct {
	mu     sync.Mutex
	path   string
	events []Event
}

func NewService() (*Service, error) {
	path, err := dataFilePath()
	if err != nil {
		return nil, err
	}

	loaded, err := load(path)
	if err != nil {
		return nil, err
	}

	return &Service{path: path, events: loaded}, nil
}

// List returns every stored event as-is (no recurrence expansion). Use
// ListOccurrences for date-ranged, recurrence-aware views.
func (s *Service) List() []Event {
	s.mu.Lock()
	defer s.mu.Unlock()

	out := make([]Event, len(s.events))
	copy(out, s.events)

	return out
}

// Add creates an event from in, after validating and normalizing it.
func (s *Service) Add(in EventInput) (Event, error) {
	in, err := in.normalized()
	if err != nil {
		return Event{}, err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	event := Event{
		ID:        uuid.NewString(),
		CreatedAt: time.Now(),
	}
	applyInput(&event, in)

	s.events = append(s.events, event)

	if err := save(s.path, s.events); err != nil {
		return Event{}, err
	}

	return event, nil
}

// Update replaces the editable fields of the event identified by id with in.
// ID, CreatedAt, Exceptions and Completions are kept as-is.
func (s *Service) Update(id string, in EventInput) (Event, error) {
	in, err := in.normalized()
	if err != nil {
		return Event{}, err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	for i := range s.events {
		if s.events[i].ID == id {
			applyInput(&s.events[i], in)

			if err := save(s.path, s.events); err != nil {
				return Event{}, err
			}

			return s.events[i], nil
		}
	}

	return Event{}, ErrNotFound
}

// applyInput copies an already-normalized EventInput onto e. A reminder
// whose lead time is unchanged keeps its existing ID, so the notifier's
// dedup key for it stays stable across edits.
func applyInput(e *Event, in EventInput) {
	e.Title = in.Title
	e.Description = in.Description
	e.Location = in.Location
	e.Start = in.Start
	e.End = in.End
	e.AllDay = in.AllDay
	e.Important = in.Important
	e.Recurrence = in.Recurrence

	existing := map[time.Duration]string{}
	for _, r := range e.Reminders {
		existing[r.LeadTime] = r.ID
	}

	reminders := []Reminder{}
	for _, seconds := range in.ReminderLeadSeconds {
		lead := time.Duration(seconds) * time.Second
		id, ok := existing[lead]
		if !ok {
			id = uuid.NewString()
		}
		delete(existing, lead) // a duplicate lead time gets a fresh ID
		reminders = append(reminders, Reminder{ID: id, LeadTime: lead})
	}
	e.Reminders = reminders
}

func (s *Service) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := range s.events {
		if s.events[i].ID == id {
			s.events = append(s.events[:i], s.events[i+1:]...)
			return save(s.path, s.events)
		}
	}

	return ErrNotFound
}

// AddReminder attaches a new Reminder with the given lead time to the event
// identified by eventID and returns the updated event.
func (s *Service) AddReminder(eventID string, leadTime time.Duration) (Event, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := range s.events {
		if s.events[i].ID == eventID {
			s.events[i].Reminders = append(s.events[i].Reminders, Reminder{
				ID:       uuid.NewString(),
				LeadTime: leadTime,
			})

			if err := save(s.path, s.events); err != nil {
				return Event{}, err
			}

			return s.events[i], nil
		}
	}

	return Event{}, ErrNotFound
}

// SkipOccurrence removes a single occurrence of a recurring event (the one
// whose raw, un-rescheduled date is originalDate) by adding a skip
// Exception, leaving the rest of the series alone.
func (s *Service) SkipOccurrence(eventID string, originalDate time.Time) (Event, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := range s.events {
		if s.events[i].ID != eventID {
			continue
		}

		s.events[i].Exceptions = append(s.events[i].Exceptions, Exception{
			OriginalDate: originalDate.In(s.events[i].Start.Location()),
			Type:         ExceptionSkip,
		})

		if err := save(s.path, s.events); err != nil {
			return Event{}, err
		}

		return s.events[i], nil
	}

	return Event{}, ErrNotFound
}

// ToggleCompletion marks the occurrence identified by (eventID,
// occurrenceDate) done if it wasn't, or un-marks it if it was — mirrors
// tasks.Service.Toggle's naming. occurrenceDate is matched by calendar date
// against each Completion, the same way Exception.OriginalDate is.
func (s *Service) ToggleCompletion(eventID string, occurrenceDate time.Time) (Event, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := range s.events {
		if s.events[i].ID != eventID {
			continue
		}

		completions := s.events[i].Completions
		matchIndex := -1
		for j, c := range completions {
			if sameCalendarDate(c.OccurrenceDate, occurrenceDate) {
				matchIndex = j
				break
			}
		}

		if matchIndex >= 0 {
			s.events[i].Completions = append(completions[:matchIndex], completions[matchIndex+1:]...)
		} else {
			s.events[i].Completions = append(completions, Completion{
				OccurrenceDate: occurrenceDate,
				CompletedAt:    time.Now(),
			})
		}

		if err := save(s.path, s.events); err != nil {
			return Event{}, err
		}

		return s.events[i], nil
	}

	return Event{}, ErrNotFound
}

// ListOccurrences expands every stored event into its occurrences within
// [rangeStart, rangeEnd], sorted by start time. This is the recurrence-aware
// read path for the widget and the month page — the frontend should never
// derive occurrence dates itself.
func (s *Service) ListOccurrences(rangeStart, rangeEnd time.Time) []OccurrenceView {
	events := s.List()

	// Always a non-nil slice: this is marshaled straight to JSON for the
	// frontend, and a nil slice becomes `null` there, not `[]` — which
	// breaks any array method the frontend calls on it (e.g. .slice()).
	out := []OccurrenceView{}
	for _, e := range events {
		for _, occ := range Occurrences(e, rangeStart, rangeEnd) {
			out = append(out, OccurrenceView{
				Occurrence:  occ,
				EventID:     e.ID,
				Title:       e.Title,
				Description: e.Description,
				Location:    e.Location,
				AllDay:      e.AllDay,
				Important:   e.Important,
				Recurring:   e.Recurrence != nil,
			})
		}
	}

	sort.Slice(out, func(i, j int) bool {
		return out[i].Start.Before(out[j].Start)
	})

	return out
}

// DueReminders returns every Reminder that is currently due (its lead time
// has elapsed but its occurrence hasn't started yet) as of now, across all
// stored events. Callers polling this for notifications are responsible for
// deduping repeat calls that return the same occurrence/reminder pair.
func (s *Service) DueReminders(now time.Time) []DueReminder {
	events := s.List()

	out := []DueReminder{} // see the same note in ListOccurrences
	for _, e := range events {
		if len(e.Reminders) == 0 {
			continue
		}

		maxLead := time.Duration(0)
		for _, r := range e.Reminders {
			if r.LeadTime > maxLead {
				maxLead = r.LeadTime
			}
		}

		for _, occ := range Occurrences(e, now, now.Add(maxLead)) {
			for _, r := range e.Reminders {
				if IsReminderDue(occ.Start, r, now) {
					out = append(out, DueReminder{
						EventID:         e.ID,
						EventTitle:      e.Title,
						ReminderID:      r.ID,
						OccurrenceStart: occ.Start,
						FireTime:        ReminderFireTime(occ.Start, r),
					})
				}
			}
		}
	}

	sort.Slice(out, func(i, j int) bool {
		return out[i].OccurrenceStart.Before(out[j].OccurrenceStart)
	})

	return out
}
