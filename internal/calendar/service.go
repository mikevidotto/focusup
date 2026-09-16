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

func (s *Service) Add(title, description, location string, start, end time.Time, allDay bool, recurrence *RecurrenceRule) (Event, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	event := Event{
		ID:          uuid.NewString(),
		Title:       title,
		Description: description,
		Location:    location,
		Start:       start,
		End:         end,
		AllDay:      allDay,
		Recurrence:  recurrence,
		CreatedAt:   time.Now(),
	}

	s.events = append(s.events, event)

	if err := save(s.path, s.events); err != nil {
		return Event{}, err
	}

	return event, nil
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

// ListOccurrences expands every stored event into its occurrences within
// [rangeStart, rangeEnd], sorted by start time. This is the recurrence-aware
// read path for the widget and the month page — the frontend should never
// derive occurrence dates itself.
func (s *Service) ListOccurrences(rangeStart, rangeEnd time.Time) []OccurrenceView {
	events := s.List()

	var out []OccurrenceView
	for _, e := range events {
		for _, occ := range Occurrences(e, rangeStart, rangeEnd) {
			out = append(out, OccurrenceView{
				Occurrence:  occ,
				EventID:     e.ID,
				Title:       e.Title,
				Description: e.Description,
				Location:    e.Location,
				AllDay:      e.AllDay,
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

	var out []DueReminder
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
