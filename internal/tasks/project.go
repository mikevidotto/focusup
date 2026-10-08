package tasks

import "time"

// A project is any outcome that takes more than one action. Tasks link to
// it by ProjectID; an active project should always have a next action.
type Project struct {
	ID          string     `json:"id"`
	Title       string     `json:"title"`
	Done        bool       `json:"done"`
	CreatedAt   time.Time  `json:"createdAt"`
	CompletedAt *time.Time `json:"completedAt"`
}
