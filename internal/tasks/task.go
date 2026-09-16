package tasks

import (
	"errors"
	"time"
)

var ErrNotFound = errors.New("task not found")

type Task struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Done      bool      `json:"done"`
	CreatedAt time.Time `json:"createdAt"`
}
