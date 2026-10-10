package jobs

import (
	"bytes"
	"encoding/csv"
	"errors"
	"io"
	"os"
	"slices"
	"strings"
)

const trackerFile = "job_search_tracker.csv"

// Application is one row of job_search_tracker.csv.
type Application struct {
	Row             int    `json:"row"`
	Date            string `json:"date"`
	Company         string `json:"company"`
	Sector          string `json:"sector"`
	Role            string `json:"role"`
	RoleType        string `json:"roleType"`
	Channel         string `json:"channel"`
	Status          string `json:"status"`
	ContactPerson   string `json:"contactPerson"`
	FitRating       string `json:"fitRating"`
	Notes           string `json:"notes"`
	CVFile          string `json:"cvFile"`
	CoverLetterFile string `json:"coverLetterFile"`
	Source          string `json:"source"`
	Deadline        string `json:"deadline"`
}

var ErrBadTracker = errors.New("tracker has no status/notes columns")

type trackerRecord struct {
	line   int // 1-based line in the file where the record starts
	fields []string
}

func readTracker(data []byte) ([]string, []trackerRecord, error) {
	reader := csv.NewReader(bytes.NewReader(data))
	reader.FieldsPerRecord = -1

	header, err := reader.Read()
	if err == io.EOF {
		return nil, nil, nil
	}
	if err != nil {
		return nil, nil, err
	}

	header = slices.Clone(header)

	var records []trackerRecord
	for {
		fields, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, nil, err
		}

		line, _ := reader.FieldPos(0)
		records = append(records, trackerRecord{line: line, fields: slices.Clone(fields)})
	}

	return header, records, nil
}

func field(header, fields []string, name string) string {
	i := slices.Index(header, name)
	if i < 0 || i >= len(fields) {
		return ""
	}

	return fields[i]
}

func toApplication(row int, header, fields []string) Application {
	get := func(name string) string { return field(header, fields, name) }

	return Application{
		Row:             row,
		Date:            get("date"),
		Company:         get("company"),
		Sector:          get("sector"),
		Role:            get("role"),
		RoleType:        get("role_type"),
		Channel:         get("channel"),
		Status:          get("status"),
		ContactPerson:   get("contact_person"),
		FitRating:       get("fit_rating"),
		Notes:           get("notes"),
		CVFile:          get("cv_file"),
		CoverLetterFile: get("cover_letter_file"),
		Source:          get("source"),
		Deadline:        get("deadline"),
	}
}

func (s *Service) ListApplications() ([]Application, error) {
	path, err := s.resolve(trackerFile)
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return []Application{}, nil
	}
	if err != nil {
		return nil, err
	}

	header, records, err := readTracker(data)
	if err != nil {
		return nil, err
	}

	apps := make([]Application, 0, len(records))
	for i, record := range records {
		apps = append(apps, toApplication(i, header, record.fields))
	}

	return apps, nil
}

// markApplied sets the record at index row to applied and notes the date,
// rewriting only that record's line so the rest of the file is untouched.
func markApplied(data []byte, row int, date string) ([]byte, Application, error) {
	header, records, err := readTracker(data)
	if err != nil {
		return nil, Application{}, err
	}

	if row < 0 || row >= len(records) {
		return nil, Application{}, ErrNotFound
	}

	statusCol := slices.Index(header, "status")
	notesCol := slices.Index(header, "notes")
	if statusCol < 0 || notesCol < 0 {
		return nil, Application{}, ErrBadTracker
	}

	record := records[row]
	fields := slices.Clone(record.fields)
	for len(fields) < len(header) {
		fields = append(fields, "")
	}

	fields[statusCol] = "applied"

	note := "applied " + date
	if strings.TrimSpace(fields[notesCol]) == "" {
		fields[notesCol] = note
	} else if !strings.Contains(fields[notesCol], note) {
		fields[notesCol] = fields[notesCol] + "; " + note
	}

	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)
	if err := writer.Write(fields); err != nil {
		return nil, Application{}, err
	}
	writer.Flush()
	encoded := strings.TrimRight(buf.String(), "\r\n")

	lines := strings.Split(string(data), "\n")
	index := record.line - 1

	// Only single-line records are rewritten in place; a quoted field with an
	// embedded newline would need the whole file re-encoded.
	if index < 0 || index >= len(lines) {
		return nil, Application{}, ErrNotFound
	}
	if row+1 < len(records) && records[row+1].line != record.line+1 {
		return nil, Application{}, ErrBadTracker
	}

	lines[index] = encoded

	return []byte(strings.Join(lines, "\n")), toApplication(row, header, fields), nil
}

// MarkApplied flips a tracker row to applied and ticks the matching apply
// queue line (matched on URL, then company + role) with the same date.
func (s *Service) MarkApplied(row int, date string) error {
	path, err := s.resolve(trackerFile)
	if err != nil {
		return err
	}

	s.mu.Lock()
	data, err := os.ReadFile(path)
	if err != nil {
		s.mu.Unlock()
		return err
	}

	updated, app, err := markApplied(data, row, date)
	if err == nil {
		err = writeFileAtomic(path, updated)
	}
	s.mu.Unlock()

	if err != nil {
		return err
	}

	queue, err := s.ListQueue()
	if err != nil {
		return err
	}

	if item, ok := matchQueue(queue, app); ok {
		return s.SetQueueOutcome(item.Number, "applied "+date)
	}

	return nil
}

func matchQueue(queue []QueueItem, app Application) (QueueItem, bool) {
	if app.Source != "" {
		for _, item := range queue {
			if item.URL == app.Source {
				return item, true
			}
		}
	}

	for _, item := range queue {
		if strings.EqualFold(item.Company, app.Company) && strings.HasPrefix(strings.ToLower(item.Role), strings.ToLower(app.Role)) {
			return item, true
		}
	}

	return QueueItem{}, false
}
