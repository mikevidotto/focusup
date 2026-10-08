package tasks

import (
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

type Service struct {
	mu           sync.Mutex
	tasksPath    string
	projectsPath string
	tasks        []Task
	projects     []Project
}

func NewService() (*Service, error) {
	dir, err := datadir()
	if err != nil {
		return nil, err
	}

	return newServiceAt(dir)
}

func newServiceAt(dir string) (*Service, error) {
	s := &Service{
		tasksPath:    filepath.Join(dir, "tasks.json"),
		projectsPath: filepath.Join(dir, "projects.json"),
	}

	var err error
	if s.tasks, err = load[Task](s.tasksPath); err != nil {
		return nil, err
	}
	if s.projects, err = load[Project](s.projectsPath); err != nil {
		return nil, err
	}

	for i := range s.tasks {
		s.tasks[i].List = normalizeList(s.tasks[i].List)
		s.tasks[i].Contexts = normalizeContexts(s.tasks[i].Contexts)
	}

	return s, nil
}

func (s *Service) List() []Task {
	s.mu.Lock()
	defer s.mu.Unlock()

	out := make([]Task, len(s.tasks))
	for i, t := range s.tasks {
		out[i] = t.clone()
	}

	return out
}

// Add captures a new task into the inbox.
func (s *Service) Add(title string, contexts []string) (Task, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.add(title, contexts, ListInbox, "")
}

// AddProjectTask adds a next action that belongs to a project.
func (s *Service) AddProjectTask(projectID, title string, contexts []string) (Task, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.projectIndex(projectID) < 0 {
		return Task{}, ErrProjectNotFound
	}

	return s.add(title, contexts, ListNext, projectID)
}

func (s *Service) add(title string, contexts []string, list, projectID string) (Task, error) {
	title = strings.TrimSpace(title)
	if title == "" {
		return Task{}, ErrEmptyTitle
	}

	task := Task{
		ID:        uuid.NewString(),
		Title:     title,
		List:      list,
		Contexts:  normalizeContexts(contexts),
		ProjectID: projectID,
		CreatedAt: time.Now(),
	}

	s.tasks = append(s.tasks, task)

	if err := save(s.tasksPath, s.tasks); err != nil {
		return Task{}, err
	}

	return task.clone(), nil
}

func (s *Service) Toggle(id string) (Task, error) {
	return s.updateTask(id, func(t *Task) error {
		t.Done = !t.Done

		if t.Done {
			now := time.Now()
			t.CompletedAt = &now
		} else {
			t.CompletedAt = nil
		}

		return nil
	})
}

// Move puts the task in another GTD list (inbox, next or someday).
func (s *Service) Move(id, list string) (Task, error) {
	if !validList(list) {
		return Task{}, ErrInvalidList
	}

	return s.updateTask(id, func(t *Task) error {
		t.List = list
		return nil
	})
}

func (s *Service) Rename(id, title string) (Task, error) {
	title = strings.TrimSpace(title)
	if title == "" {
		return Task{}, ErrEmptyTitle
	}

	return s.updateTask(id, func(t *Task) error {
		t.Title = title
		return nil
	})
}

func (s *Service) SetContexts(id string, contexts []string) (Task, error) {
	return s.updateTask(id, func(t *Task) error {
		t.Contexts = normalizeContexts(contexts)
		return nil
	})
}

// SetProject links the task to a project, or unlinks it when projectID is
// empty. Assigning a project clarifies an inbox item into a next action.
func (s *Service) SetProject(id, projectID string) (Task, error) {
	return s.updateTask(id, func(t *Task) error {
		if projectID != "" && s.projectIndex(projectID) < 0 {
			return ErrProjectNotFound
		}

		t.ProjectID = projectID
		if projectID != "" && t.List == ListInbox {
			t.List = ListNext
		}

		return nil
	})
}

func (s *Service) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	i := s.taskIndex(id)
	if i < 0 {
		return ErrNotFound
	}

	s.tasks = slices.Delete(s.tasks, i, i+1)
	return save(s.tasksPath, s.tasks)
}

func (s *Service) ListProjects() []Project {
	s.mu.Lock()
	defer s.mu.Unlock()

	return slices.Clone(s.projects)
}

func (s *Service) AddProject(title string) (Project, error) {
	title = strings.TrimSpace(title)
	if title == "" {
		return Project{}, ErrEmptyTitle
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	project := Project{
		ID:        uuid.NewString(),
		Title:     title,
		CreatedAt: time.Now(),
	}

	s.projects = append(s.projects, project)

	if err := save(s.projectsPath, s.projects); err != nil {
		return Project{}, err
	}

	return project, nil
}

func (s *Service) RenameProject(id, title string) (Project, error) {
	title = strings.TrimSpace(title)
	if title == "" {
		return Project{}, ErrEmptyTitle
	}

	return s.updateProject(id, func(p *Project) {
		p.Title = title
	})
}

func (s *Service) ToggleProject(id string) (Project, error) {
	return s.updateProject(id, func(p *Project) {
		p.Done = !p.Done

		if p.Done {
			now := time.Now()
			p.CompletedAt = &now
		} else {
			p.CompletedAt = nil
		}
	})
}

// DeleteProject removes the project and unlinks its tasks; the tasks
// themselves are kept.
func (s *Service) DeleteProject(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	i := s.projectIndex(id)
	if i < 0 {
		return ErrProjectNotFound
	}

	unlinked := false
	for j := range s.tasks {
		if s.tasks[j].ProjectID == id {
			s.tasks[j].ProjectID = ""
			unlinked = true
		}
	}

	if unlinked {
		if err := save(s.tasksPath, s.tasks); err != nil {
			return err
		}
	}

	s.projects = slices.Delete(s.projects, i, i+1)
	return save(s.projectsPath, s.projects)
}

func (s *Service) updateTask(id string, apply func(*Task) error) (Task, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	i := s.taskIndex(id)
	if i < 0 {
		return Task{}, ErrNotFound
	}

	updated := s.tasks[i].clone()
	if err := apply(&updated); err != nil {
		return Task{}, err
	}
	s.tasks[i] = updated

	if err := save(s.tasksPath, s.tasks); err != nil {
		return Task{}, err
	}

	return updated.clone(), nil
}

func (s *Service) updateProject(id string, apply func(*Project)) (Project, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	i := s.projectIndex(id)
	if i < 0 {
		return Project{}, ErrProjectNotFound
	}

	apply(&s.projects[i])

	if err := save(s.projectsPath, s.projects); err != nil {
		return Project{}, err
	}

	return s.projects[i], nil
}

func (s *Service) taskIndex(id string) int {
	return slices.IndexFunc(s.tasks, func(t Task) bool { return t.ID == id })
}

func (s *Service) projectIndex(id string) int {
	return slices.IndexFunc(s.projects, func(p Project) bool { return p.ID == id })
}
