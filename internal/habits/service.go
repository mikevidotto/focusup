package habits

import (
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

type Service struct {
	mu     sync.Mutex
	path   string
	habits []Habit
}

func NewService() (*Service, error) {
	path, err := datafilepath()
	if err != nil {
		return nil, err
	}

	loaded, err := load(path)
	if err != nil {
		return nil, err
	}

	return &Service{path: path, habits: loaded}, nil
}

func (s *Service) List() []Habit {
	s.mu.Lock()
	defer s.mu.Unlock()

	out := make([]Habit, len(s.habits))
	for i, h := range s.habits {
		out[i] = h.clone()
	}

	return out
}

func (s *Service) Add(name string) (Habit, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return Habit{}, ErrEmptyName
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	habit := Habit{
		ID:          uuid.NewString(),
		Name:        name,
		CreatedAt:   time.Now(),
		Completions: []string{},
	}

	s.habits = append(s.habits, habit)

	if err := save(s.path, s.habits); err != nil {
		return Habit{}, err
	}

	return habit.clone(), nil
}

func (s *Service) Rename(id, name string) (Habit, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return Habit{}, ErrEmptyName
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	i := s.indexOf(id)
	if i < 0 {
		return Habit{}, ErrNotFound
	}

	s.habits[i].Name = name

	if err := save(s.path, s.habits); err != nil {
		return Habit{}, err
	}

	return s.habits[i].clone(), nil
}

// ToggleCompletion marks the habit done on the given day (YYYY-MM-DD), or
// un-marks it if it was already done that day.
func (s *Service) ToggleCompletion(id, date string) (Habit, error) {
	if _, err := time.Parse(DateLayout, date); err != nil {
		return Habit{}, ErrInvalidDate
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	i := s.indexOf(id)
	if i < 0 {
		return Habit{}, ErrNotFound
	}

	completions := s.habits[i].Completions
	if j := slices.Index(completions, date); j >= 0 {
		completions = slices.Delete(completions, j, j+1)
	} else {
		completions = append(completions, date)
		slices.Sort(completions)
	}
	s.habits[i].Completions = completions

	if err := save(s.path, s.habits); err != nil {
		return Habit{}, err
	}

	return s.habits[i].clone(), nil
}

func (s *Service) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	i := s.indexOf(id)
	if i < 0 {
		return ErrNotFound
	}

	s.habits = slices.Delete(s.habits, i, i+1)
	return save(s.path, s.habits)
}

func (s *Service) indexOf(id string) int {
	return slices.IndexFunc(s.habits, func(h Habit) bool { return h.ID == id })
}

// clone copies the Completions slice so callers can't mutate service state.
func (h Habit) clone() Habit {
	h.Completions = slices.Clone(h.Completions)
	if h.Completions == nil {
		h.Completions = []string{}
	}
	return h
}
