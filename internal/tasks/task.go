package tasks

import (
	"errors"
	"slices"
	"strings"
	"time"
)

var (
	ErrNotFound        = errors.New("task not found")
	ErrEmptyTitle      = errors.New("title cannot be empty")
	ErrInvalidList     = errors.New("invalid list, expected inbox, next or someday")
	ErrProjectNotFound = errors.New("project not found")
)

// GTD lists a task can live in. Everything is captured into the inbox,
// then clarified into next actions or parked in someday/maybe.
const (
	ListInbox   = "inbox"
	ListNext    = "next"
	ListSomeday = "someday"
)

func validList(list string) bool {
	switch list {
	case ListInbox, ListNext, ListSomeday:
		return true
	default:
		return false
	}
}

// normalizeList defaults unknown lists to next actions: tasks saved before
// the GTD lists existed were already-clarified todos, not inbox items.
func normalizeList(list string) string {
	if validList(list) {
		return list
	}
	return ListNext
}

// normalizeContexts lowercases contexts, strips a leading "@", and drops
// blanks and duplicates, returning them sorted.
func normalizeContexts(contexts []string) []string {
	out := []string{}
	for _, c := range contexts {
		c = strings.ToLower(strings.TrimPrefix(strings.TrimSpace(c), "@"))
		if c != "" && !slices.Contains(out, c) {
			out = append(out, c)
		}
	}
	slices.Sort(out)
	return out
}

type Task struct {
	ID          string     `json:"id"`
	Title       string     `json:"title"`
	Done        bool       `json:"done"`
	List        string     `json:"list"`
	Contexts    []string   `json:"contexts"`
	ProjectID   string     `json:"projectId"`
	CreatedAt   time.Time  `json:"createdAt"`
	CompletedAt *time.Time `json:"completedAt"`
}

// clone copies the Contexts slice so callers can't mutate service state.
func (t Task) clone() Task {
	t.Contexts = slices.Clone(t.Contexts)
	if t.Contexts == nil {
		t.Contexts = []string{}
	}
	return t
}
