package habits

import (
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
	copy(out, s.habits)

	return out
}

func (s *Service) Add(name string) (Habit, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	habit := Habit{
		ID:        uuid.NewString(),
		Name:      name,
		Done:      false,
		CreatedAt: time.Now(),
	}

	s.habits = append(s.habits, habit)

	if err := save(s.path, s.habits); err != nil {
		return Habit{}, err
	}

	return habit, nil
}

func (s *Service) Toggle(id string) (Habit, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := range s.habits {
		if s.habits[i].ID == id {
			s.habits[i].Done = !s.habits[i].Done

			if s.habits[i].Done {
				now := time.Now()
				s.habits[i].CompletedAt = &now
			} else {
				s.habits[i].CompletedAt = nil
			}

			if err := save(s.path, s.habits); err != nil {
				return Habit{}, err
			}

			return s.habits[i], nil
		}
	}

	return Habit{}, ErrNotFound
}

func (s *Service) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := range s.habits {
		if s.habits[i].ID == id {
			s.habits = append(s.habits[:i], s.habits[i+1:]...)
			return save(s.path, s.habits)
		}
	}

	return ErrNotFound
}
