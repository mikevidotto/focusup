package tasks

import (
	"errors"
	"time"
)

var ErrNotFound = errors.New("task not found")

const (
	PriorityLow    = "low"
	PriorityMedium = "medium"
	PriorityHigh   = "high"
)

func normalizePriority(priority string) string {
	switch priority {
	case PriorityLow, PriorityMedium, PriorityHigh:
		return priority
	default:
		return PriorityMedium
	}
}

type Task struct {
	ID          string     `json:"id"`
	Title       string     `json:"title"`
	Done        bool       `json:"done"`
	Priority    string     `json:"priority"`
	CreatedAt   time.Time  `json:"createdAt"`
	CompletedAt *time.Time `json:"completedAt"`
}
