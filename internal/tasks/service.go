package tasks

import (
	"sync"
	"time"

	"github.com/google/uuid"
)

type Service struct {
	mu    sync.Mutex
	path  string
	tasks []Task
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

	return &Service{path: path, tasks: loaded}, nil
}

func (s *Service) List() []Task {
	s.mu.Lock()
	defer s.mu.Unlock()

	out := make([]Task, len(s.tasks))
	copy(out, s.tasks)

	return out
}

func (s *Service) Add(title string) (Task, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	task := Task{
		ID:        uuid.NewString(),
		Title:     title,
		CreatedAt: time.Now(),
	}

	s.tasks = append(s.tasks, task)

	if err := save(s.path, s.tasks); err != nil {
		return Task{}, err
	}

	return task, nil
}

func (s *Service) Toggle(id string) (Task, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := range s.tasks {
		if s.tasks[i].ID == id {
			s.tasks[i].Done = !s.tasks[i].Done

			if err := save(s.path, s.tasks); err != nil {
				return Task{}, err
			}

			return s.tasks[i], nil
		}
	}

	return Task{}, ErrNotFound
}

func (s *Service) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := range s.tasks {
		if s.tasks[i].ID == id {
			s.tasks = append(s.tasks[:i], s.tasks[i+1:]...)
			return save(s.path, s.tasks)
		}
	}

	return ErrNotFound
}
