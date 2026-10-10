package journal

import (
	"slices"
	"strings"
	"sync"
	"time"
)

type Service struct {
	mu      sync.Mutex
	path    string
	entries []Entry
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

	s := &Service{path: path, entries: loaded}
	s.sort()

	return s, nil
}

// List returns every entry, newest date first.
func (s *Service) List() []Entry {
	s.mu.Lock()
	defer s.mu.Unlock()

	out := make([]Entry, len(s.entries))
	for i, e := range s.entries {
		out[i] = e.clone()
	}

	return out
}

func (s *Service) Get(date string) (Entry, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	i := s.indexOf(date)
	if i < 0 {
		return Entry{}, ErrNotFound
	}

	return s.entries[i].clone(), nil
}

// Save creates or replaces the entry for date (YYYY-MM-DD). Saving an entry
// with nothing written in it removes it, so opening a day and leaving it
// blank doesn't leave an empty record behind.
func (s *Service) Save(date string, prompts []PromptAnswer, body string) (Entry, error) {
	if _, err := time.Parse(DateLayout, date); err != nil {
		return Entry{}, ErrInvalidDate
	}

	entry := Entry{
		Date:      date,
		Prompts:   make([]PromptAnswer, 0, len(prompts)),
		Body:      strings.TrimRight(body, " \t\r\n"),
		UpdatedAt: time.Now(),
	}
	for _, p := range prompts {
		entry.Prompts = append(entry.Prompts, PromptAnswer{
			Prompt: strings.TrimSpace(p.Prompt),
			Answer: strings.TrimRight(p.Answer, " \t\r\n"),
		})
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	i := s.indexOf(date)

	if entry.isBlank() {
		if i >= 0 {
			s.entries = slices.Delete(s.entries, i, i+1)
			if err := save(s.path, s.entries); err != nil {
				return Entry{}, err
			}
		}
		return entry.clone(), nil
	}

	if i >= 0 {
		s.entries[i] = entry
	} else {
		s.entries = append(s.entries, entry)
		s.sort()
	}

	if err := save(s.path, s.entries); err != nil {
		return Entry{}, err
	}

	return entry.clone(), nil
}

func (s *Service) Delete(date string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	i := s.indexOf(date)
	if i < 0 {
		return ErrNotFound
	}

	s.entries = slices.Delete(s.entries, i, i+1)
	return save(s.path, s.entries)
}

func (s *Service) indexOf(date string) int {
	return slices.IndexFunc(s.entries, func(e Entry) bool { return e.Date == date })
}

// sort keeps entries newest-first; YYYY-MM-DD strings sort chronologically.
func (s *Service) sort() {
	slices.SortFunc(s.entries, func(a, b Entry) int { return strings.Compare(b.Date, a.Date) })
}

func (e Entry) isBlank() bool {
	if strings.TrimSpace(e.Body) != "" {
		return false
	}
	for _, p := range e.Prompts {
		if strings.TrimSpace(p.Answer) != "" {
			return false
		}
	}
	return true
}

// clone copies the Prompts slice so callers can't mutate service state.
func (e Entry) clone() Entry {
	e.Prompts = slices.Clone(e.Prompts)
	if e.Prompts == nil {
		e.Prompts = []PromptAnswer{}
	}
	return e
}
