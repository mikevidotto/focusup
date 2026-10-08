package workouts

import (
	"errors"
	"math"
	"time"
)

var (
	ErrNotFound      = errors.New("cycle or workout not found")
	ErrOutOfOrder    = errors.New("workouts must be logged in order")
	ErrInvalidDate   = errors.New("invalid date, expected YYYY-MM-DD in program order")
	ErrInvalidInput  = errors.New("every lift needs a weight greater than 0")
	ErrAlreadySetUp  = errors.New("program is already set up")
	ErrNothingToUndo = errors.New("no logged workouts to undo")
)

// DateLayout matches habits.DateLayout: plain local calendar days.
const DateLayout = "2006-01-02"

// Lifts in the order they are trained within a week (normally Mon, Tue,
// Thu, Fri). A cycle is 4 weeks × 4 lifts = 16 workouts, addressed by
// index (week-1)*4 + liftIndex.
var Lifts = []string{"squat", "bench", "deadlift", "press"}

const (
	WeeksPerCycle    = 4
	WorkoutsPerCycle = WeeksPerCycle * 4
	DeloadWeek       = 4

	// Days between the last workout of a cycle (press) and the next
	// cycle's first (squat): Fri -> Mon.
	DaysBetweenCycles = 3
)

// TMIncrement is how much each lift's training max goes up per cycle:
// lower-body lifts +10 lb, upper-body +5 lb.
var TMIncrement = map[string]float64{
	"squat":    10,
	"bench":    5,
	"deadlift": 10,
	"press":    5,
}

type Cycle struct {
	ID          string             `json:"id"`
	Number      int                `json:"number"`
	StartDate   string             `json:"startDate"`
	TrainingMax map[string]float64 `json:"trainingMax"`
	Logs        []WorkoutLog       `json:"logs"`
	CreatedAt   time.Time          `json:"createdAt"`
}

// WorkoutLog records a workout actually done. Logs are kept in index
// order, so Logs[i].Index == i.
type WorkoutLog struct {
	Index     int    `json:"index"`
	Date      string `json:"date"`
	AmrapReps int    `json:"amrapReps"`
}

func Round5(weight float64) float64 {
	return math.Round(weight/5) * 5
}

func TrainingMaxFrom1RM(oneRepMax float64) float64 {
	return Round5(0.9 * oneRepMax)
}

func NextTrainingMaxes(prev map[string]float64) map[string]float64 {
	next := make(map[string]float64, len(Lifts))
	for _, lift := range Lifts {
		next[lift] = prev[lift] + TMIncrement[lift]
	}
	return next
}

func weekOf(index int) int {
	return index/len(Lifts) + 1
}

func validWeights(weights map[string]float64) bool {
	for _, lift := range Lifts {
		if weights[lift] <= 0 {
			return false
		}
	}
	return true
}

func parseDate(date string) (time.Time, error) {
	t, err := time.Parse(DateLayout, date)
	if err != nil {
		return time.Time{}, ErrInvalidDate
	}
	return t, nil
}

func (c Cycle) clone() Cycle {
	tm := make(map[string]float64, len(c.TrainingMax))
	for k, v := range c.TrainingMax {
		tm[k] = v
	}
	c.TrainingMax = tm
	c.Logs = append([]WorkoutLog{}, c.Logs...)
	return c
}
