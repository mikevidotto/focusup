package habits

import (
	"errors"
	"time"
)

var (
	ErrNotFound    = errors.New("habit not found")
	ErrEmptyName   = errors.New("habit name cannot be empty")
	ErrInvalidDate = errors.New("invalid date, expected YYYY-MM-DD")
)

// DateLayout is the format for completion dates. Dates are plain calendar
// days (no time or zone) so the frontend's local "today" always maps to
// the same box regardless of the machine's UTC offset.
const DateLayout = "2006-01-02"

/*
A habit is a single recurring item (e.g. "Wake up at 5:30"). Rather than
creating a new record every day, each habit keeps the set of days it was
completed on, so any week (current or past) can be rendered by checking
which of its seven dates appear in Completions.
*/
type Habit struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	CreatedAt   time.Time `json:"createdAt"`
	Completions []string  `json:"completions"`
}
