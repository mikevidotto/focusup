package settings

import (
	"encoding/json"
	"errors"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

var (
	ErrInvalidTheme     = errors.New("theme must be \"dark\" or \"light\"")
	ErrInvalidServerURL = errors.New("workout server URL must start with http:// or https://")
)

type Settings struct {
	Theme string `json:"theme"`
	// JobSearchDir is the ai-job-search repo the Jobs tab reads; empty means
	// the default location (see jobs.DefaultDir).
	JobSearchDir string `json:"jobSearchDir"`
	// LastReviewAt is when the GTD weekly review was last finished; nil
	// means never.
	LastReviewAt *time.Time `json:"lastReviewAt"`
	// WorkoutServerURL is the gym server (cmd/gymserver) that holds the
	// 5/3/1 program; empty keeps it in the local data file.
	WorkoutServerURL string `json:"workoutServerURL"`
}

func defaults() Settings {
	return Settings{Theme: "dark"}
}

type Service struct {
	mu       sync.Mutex
	path     string
	settings Settings
}

func NewService() (*Service, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return nil, err
	}

	appdir := filepath.Join(dir, "focusup")
	if err := os.MkdirAll(appdir, 0o755); err != nil {
		return nil, err
	}

	path := filepath.Join(appdir, "settings.json")
	loaded, err := load(path)
	if err != nil {
		return nil, err
	}

	return &Service{path: path, settings: loaded}, nil
}

func (s *Service) Get() Settings {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.settings
}

func (s *Service) SetTheme(theme string) (Settings, error) {
	if theme != "dark" && theme != "light" {
		return Settings{}, ErrInvalidTheme
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.settings.Theme = theme
	if err := save(s.path, s.settings); err != nil {
		return Settings{}, err
	}

	return s.settings, nil
}

// SetJobSearchDir persists the ai-job-search repo path ("" resets it to the
// default).
func (s *Service) SetJobSearchDir(dir string) (Settings, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.settings.JobSearchDir = strings.TrimSpace(dir)
	if err := save(s.path, s.settings); err != nil {
		return Settings{}, err
	}

	return s.settings, nil
}

// SetWorkoutServerURL points the Workouts page at a gym server ("" goes
// back to the local data file).
func (s *Service) SetWorkoutServerURL(raw string) (Settings, error) {
	raw = strings.TrimRight(strings.TrimSpace(raw), "/")
	if raw != "" {
		u, err := url.Parse(raw)
		if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
			return Settings{}, ErrInvalidServerURL
		}
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.settings.WorkoutServerURL = raw
	if err := save(s.path, s.settings); err != nil {
		return Settings{}, err
	}

	return s.settings, nil
}

// MarkReviewed records that the weekly review was just finished.
func (s *Service) MarkReviewed() (Settings, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	s.settings.LastReviewAt = &now
	if err := save(s.path, s.settings); err != nil {
		return Settings{}, err
	}

	return s.settings, nil
}

func load(path string) (Settings, error) {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return defaults(), nil
	}
	if err != nil {
		return Settings{}, err
	}

	loaded := defaults()
	if err := json.Unmarshal(data, &loaded); err != nil {
		return Settings{}, err
	}

	return loaded, nil
}

func save(path string, settings Settings) error {
	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return err
	}

	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}

	return os.Rename(tmp, path)
}
