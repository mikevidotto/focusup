package workouts

import (
	"sync"
	"time"

	"github.com/google/uuid"
)

// Service stores the 5/3/1 program as a list of cycles, oldest first. Only
// the latest cycle accepts new logs; the next cycle is generated
// automatically when its 16th workout is logged.
type Service struct {
	mu     sync.Mutex
	path   string
	cycles []Cycle
}

func NewService() (*Service, error) {
	path, err := datafilepath()
	if err != nil {
		return nil, err
	}

	loaded, err := load(path)
	if err != nil {
		return nil, err
	}

	return &Service{path: path, cycles: loaded}, nil
}

func (s *Service) ListCycles() []Cycle {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.snapshot()
}

// Setup creates cycle 1 from the lifter's true 1RMs (TM = 90%, rounded to 5).
func (s *Service) Setup(startDate string, oneRepMax map[string]float64) (Cycle, error) {
	if _, err := parseDate(startDate); err != nil {
		return Cycle{}, err
	}
	if !validWeights(oneRepMax) {
		return Cycle{}, ErrInvalidInput
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if len(s.cycles) > 0 {
		return Cycle{}, ErrAlreadySetUp
	}

	tm := make(map[string]float64, len(Lifts))
	for _, lift := range Lifts {
		tm[lift] = TrainingMaxFrom1RM(oneRepMax[lift])
	}

	cycle := newCycle(1, startDate, tm)
	s.cycles = []Cycle{cycle}

	if err := save(s.path, s.cycles); err != nil {
		return Cycle{}, err
	}
	return cycle.clone(), nil
}

// CompleteWorkout logs the next workout of the latest cycle as done on
// date. Logging the cycle's last workout generates the next cycle.
func (s *Service) CompleteWorkout(cycleID string, index int, date string, amrapReps int) ([]Cycle, error) {
	if _, err := parseDate(date); err != nil {
		return nil, err
	}
	if amrapReps < 0 {
		return nil, ErrInvalidInput
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	ci := len(s.cycles) - 1
	if ci < 0 || s.cycles[ci].ID != cycleID {
		return nil, ErrNotFound
	}
	cycle := &s.cycles[ci]

	if index != len(cycle.Logs) || index >= WorkoutsPerCycle {
		return nil, ErrOutOfOrder
	}
	if prev, ok := s.prevLogDate(ci, index); ok && date < prev {
		return nil, ErrInvalidDate
	}

	if weekOf(index) == DeloadWeek {
		amrapReps = 0
	}

	// Starting earlier than planned just moves the cycle's start date.
	if index == 0 && date < cycle.StartDate {
		cycle.StartDate = date
	}

	cycle.Logs = append(cycle.Logs, WorkoutLog{Index: index, Date: date, AmrapReps: amrapReps})

	if len(cycle.Logs) == WorkoutsPerCycle {
		start, _ := parseDate(date)
		next := newCycle(
			cycle.Number+1,
			start.AddDate(0, 0, DaysBetweenCycles).Format(DateLayout),
			NextTrainingMaxes(cycle.TrainingMax),
		)
		s.cycles = append(s.cycles, next)
	}

	if err := save(s.path, s.cycles); err != nil {
		return nil, err
	}
	return s.snapshot(), nil
}

// UpdateCycle changes a cycle's planned start date and training maxes.
func (s *Service) UpdateCycle(cycleID, startDate string, trainingMax map[string]float64) (Cycle, error) {
	if _, err := parseDate(startDate); err != nil {
		return Cycle{}, err
	}
	if !validWeights(trainingMax) {
		return Cycle{}, ErrInvalidInput
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	ci := s.indexOf(cycleID)
	if ci < 0 {
		return Cycle{}, ErrNotFound
	}
	cycle := &s.cycles[ci]

	if len(cycle.Logs) > 0 && startDate > cycle.Logs[0].Date {
		return Cycle{}, ErrInvalidDate
	}
	if prev, ok := s.prevLogDate(ci, 0); ok && startDate < prev {
		return Cycle{}, ErrInvalidDate
	}

	cycle.StartDate = startDate
	for _, lift := range Lifts {
		cycle.TrainingMax[lift] = trainingMax[lift]
	}

	if err := save(s.path, s.cycles); err != nil {
		return Cycle{}, err
	}
	return cycle.clone(), nil
}

// UpdateWorkoutLog corrects the date or AMRAP reps of a logged workout.
// The date has to stay between the neighbouring logged workouts.
func (s *Service) UpdateWorkoutLog(cycleID string, index int, date string, amrapReps int) (Cycle, error) {
	if _, err := parseDate(date); err != nil {
		return Cycle{}, err
	}
	if amrapReps < 0 {
		return Cycle{}, ErrInvalidInput
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	ci := s.indexOf(cycleID)
	if ci < 0 || index < 0 || index >= len(s.cycles[ci].Logs) {
		return Cycle{}, ErrNotFound
	}
	cycle := &s.cycles[ci]

	if prev, ok := s.prevLogDate(ci, index); ok && date < prev {
		return Cycle{}, ErrInvalidDate
	}
	if next, ok := s.nextLogDate(ci, index); ok && date > next {
		return Cycle{}, ErrInvalidDate
	}

	if weekOf(index) == DeloadWeek {
		amrapReps = 0
	}
	if index == 0 && date < cycle.StartDate {
		cycle.StartDate = date
	}

	cycle.Logs[index].Date = date
	cycle.Logs[index].AmrapReps = amrapReps

	if err := save(s.path, s.cycles); err != nil {
		return Cycle{}, err
	}
	return cycle.clone(), nil
}

// UndoLastWorkout removes the most recent log. If that log finished a
// cycle, the cycle that was generated from it is dropped too.
func (s *Service) UndoLastWorkout() ([]Cycle, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	ci := len(s.cycles) - 1
	if ci < 0 {
		return nil, ErrNothingToUndo
	}

	if len(s.cycles[ci].Logs) == 0 {
		if ci == 0 {
			return nil, ErrNothingToUndo
		}
		s.cycles = s.cycles[:ci]
		ci--
	}

	cycle := &s.cycles[ci]
	cycle.Logs = cycle.Logs[:len(cycle.Logs)-1]

	if err := save(s.path, s.cycles); err != nil {
		return nil, err
	}
	return s.snapshot(), nil
}

// Reset deletes the whole program so setup can start over.
func (s *Service) Reset() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.cycles = []Cycle{}
	return save(s.path, s.cycles)
}

func newCycle(number int, startDate string, trainingMax map[string]float64) Cycle {
	return Cycle{
		ID:          uuid.NewString(),
		Number:      number,
		StartDate:   startDate,
		TrainingMax: trainingMax,
		Logs:        []WorkoutLog{},
		CreatedAt:   time.Now(),
	}
}

func (s *Service) indexOf(cycleID string) int {
	for i := range s.cycles {
		if s.cycles[i].ID == cycleID {
			return i
		}
	}
	return -1
}

// prevLogDate is the date of the log just before (ci, index), which may be
// the previous cycle's last workout.
func (s *Service) prevLogDate(ci, index int) (string, bool) {
	if index > 0 {
		return s.cycles[ci].Logs[index-1].Date, true
	}
	if ci > 0 {
		logs := s.cycles[ci-1].Logs
		if len(logs) > 0 {
			return logs[len(logs)-1].Date, true
		}
	}
	return "", false
}

// nextLogDate is the date of the log just after (ci, index), which may be
// the next cycle's first workout.
func (s *Service) nextLogDate(ci, index int) (string, bool) {
	if index+1 < len(s.cycles[ci].Logs) {
		return s.cycles[ci].Logs[index+1].Date, true
	}
	if ci+1 < len(s.cycles) && len(s.cycles[ci+1].Logs) > 0 {
		return s.cycles[ci+1].Logs[0].Date, true
	}
	return "", false
}

func (s *Service) snapshot() []Cycle {
	out := make([]Cycle, len(s.cycles))
	for i, c := range s.cycles {
		out[i] = c.clone()
	}
	return out
}
