package habits

import (
	"time"
    "errors"
)

var ErrNotFound = errors.New("habit not found")

/*
habits will be named
habits can be checked and unchecked
instances of a habit exist each day
*/
type Habit struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Done        bool       `json:"done"`
	CreatedAt   time.Time  `json:"createdAT"`
	CompletedAt *time.Time `json:"dateCompleted"`
}
