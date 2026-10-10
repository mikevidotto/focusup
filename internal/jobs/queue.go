package jobs

import (
	"os"
	"regexp"
	"strconv"
	"strings"
)

const queueFile = "job_scraper/apply_queue.md"

// QueueItem is one row of the apply queue table:
//
//	| # | Done | Fit | Role | Company | Notes | URL | Outcome |
type QueueItem struct {
	Number  int    `json:"number"`
	Done    bool   `json:"done"`
	Fit     string `json:"fit"`
	Role    string `json:"role"`
	Company string `json:"company"`
	Notes   string `json:"notes"`
	URL     string `json:"url"`
	Outcome string `json:"outcome"`
}

const queueColumns = 8

var queueRow = regexp.MustCompile(`^\|\s*\d+\s*\|`)

func splitRow(line string) []string {
	line = strings.TrimSpace(line)
	line = strings.TrimPrefix(line, "|")
	line = strings.TrimSuffix(line, "|")

	cells := strings.Split(line, "|")
	for i, cell := range cells {
		cells[i] = strings.TrimSpace(cell)
	}

	return cells
}

func parseQueueLine(line string) (QueueItem, bool) {
	if !queueRow.MatchString(line) {
		return QueueItem{}, false
	}

	cells := splitRow(line)
	if len(cells) != queueColumns {
		return QueueItem{}, false
	}

	number, err := strconv.Atoi(cells[0])
	if err != nil {
		return QueueItem{}, false
	}

	return QueueItem{
		Number:  number,
		Done:    strings.EqualFold(cells[1], "[x]"),
		Fit:     cells[2],
		Role:    cells[3],
		Company: cells[4],
		Notes:   cells[5],
		URL:     cells[6],
		Outcome: cells[7],
	}, true
}

func parseQueue(text string) []QueueItem {
	items := []QueueItem{}

	for _, line := range strings.Split(text, "\n") {
		if item, ok := parseQueueLine(line); ok {
			items = append(items, item)
		}
	}

	return items
}

// cleanCell keeps a value from breaking the markdown table.
func cleanCell(value string) string {
	value = strings.NewReplacer("|", "/", "\r", " ", "\n", " ").Replace(value)
	return strings.TrimSpace(value)
}

// setQueueOutcome ticks row number and writes its outcome, leaving every
// other line of text untouched.
func setQueueOutcome(text string, number int, outcome string) (string, error) {
	lines := strings.Split(text, "\n")

	for i, line := range lines {
		item, ok := parseQueueLine(line)
		if !ok || item.Number != number {
			continue
		}

		cells := splitRow(line)
		cells[1] = "[x]"
		cells[7] = cleanCell(outcome)
		lines[i] = "| " + strings.Join(cells, " | ") + " |"

		return strings.Join(lines, "\n"), nil
	}

	return "", ErrNotFound
}

func (s *Service) ListQueue() ([]QueueItem, error) {
	path, err := s.resolve(queueFile)
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return []QueueItem{}, nil
	}
	if err != nil {
		return nil, err
	}

	return parseQueue(string(data)), nil
}

// SetQueueOutcome ticks a queue row and records its outcome, e.g.
// "skipped - needs a bachelor's degree" or "applied 2026-10-08".
func (s *Service) SetQueueOutcome(number int, outcome string) error {
	path, err := s.resolve(queueFile)
	if err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	updated, err := setQueueOutcome(string(data), number, outcome)
	if err != nil {
		return err
	}

	return writeFileAtomic(path, []byte(updated))
}
