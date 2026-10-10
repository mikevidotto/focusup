package journal

import (
	"errors"
	"time"
)

var (
	ErrNotFound    = errors.New("journal entry not found")
	ErrInvalidDate = errors.New("invalid date, expected YYYY-MM-DD")
)

// DateLayout is the format for entry dates. Like habits, an entry belongs to
// a plain calendar day (no time or zone) so the frontend's local "today"
// always maps to the same entry regardless of the machine's UTC offset.
const DateLayout = "2006-01-02"

// PromptAnswer pairs a prompt with what was written for it. The prompt text
// is stored alongside the answer so past entries keep their original prompts
// even if the frontend's prompt pool changes.
type PromptAnswer struct {
	Prompt string `json:"prompt"`
	Answer string `json:"answer"`
}

/*
An Entry is the journal page for a single day: a handful of answered
prompts plus a free-write body. There is at most one entry per date.
*/
type Entry struct {
	Date      string         `json:"date"`
	Prompts   []PromptAnswer `json:"prompts"`
	Body      string         `json:"body"`
	UpdatedAt time.Time      `json:"updatedAt"`
}
