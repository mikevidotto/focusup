// Package jobs connects FocusUp to an ai-job-search workspace: it reads the
// apply queue and application tracker that live in that repo, and records
// when a job is applied to or skipped. Scraping and drafting happen outside
// FocusUp; this package only tracks the results.
//
// The repo stays the single source of truth; nothing is copied into
// FocusUp's own config dir.
package jobs

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

var (
	ErrNoDir      = errors.New("job search folder is not set")
	ErrOutsideDir = errors.New("path is outside the job search folder")
	ErrNotFound   = errors.New("not found")
)

type Service struct {
	mu  sync.Mutex
	dir string
}

// DefaultDir is where the ai-job-search repo is expected when the setting is
// empty.
func DefaultDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}

	return filepath.Join(home, "dev", "projects", "ai-job-search")
}

func NewService(dir string) *Service {
	if strings.TrimSpace(dir) == "" {
		dir = DefaultDir()
	}

	return &Service{dir: dir}
}

func (s *Service) Dir() string {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.dir
}

func (s *Service) SetDir(dir string) {
	if strings.TrimSpace(dir) == "" {
		dir = DefaultDir()
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.dir = dir
}

// resolve turns a repo-relative path into an absolute one, refusing anything
// that would land outside the repo.
func (s *Service) resolve(rel string) (string, error) {
	dir := s.Dir()
	if dir == "" {
		return "", ErrNoDir
	}

	if rel == "" || filepath.IsAbs(rel) {
		return "", ErrOutsideDir
	}

	full := filepath.Join(dir, rel)

	back, err := filepath.Rel(dir, full)
	if err != nil || back == ".." || strings.HasPrefix(back, ".."+string(filepath.Separator)) {
		return "", ErrOutsideDir
	}

	return full, nil
}

// writeFileAtomic mirrors the tmp + rename pattern the other stores use.
func writeFileAtomic(path string, data []byte) error {
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}

	return os.Rename(tmp, path)
}
