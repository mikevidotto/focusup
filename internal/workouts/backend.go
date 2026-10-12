package workouts

// Backend is everything the desktop app needs from the program. Service
// keeps it in a local file; Client keeps it on the gym server (see
// cmd/gymserver) so the phone and the desktop share one copy.
type Backend interface {
	ListCycles() []Cycle
	Setup(startDate string, oneRepMax map[string]float64) (Cycle, error)
	CompleteWorkout(cycleID string, index int, date string, amrapReps int) ([]Cycle, error)
	UpdateCycle(cycleID, startDate string, trainingMax map[string]float64) (Cycle, error)
	UpdateWorkoutLog(cycleID string, index int, date string, amrapReps int) (Cycle, error)
	UndoLastWorkout() ([]Cycle, error)
	Reset() error
}

var (
	_ Backend = (*Service)(nil)
	_ Backend = (*Client)(nil)
)

// Request bodies for the HTTP API, shared by Handler and Client.

type setupRequest struct {
	StartDate string             `json:"startDate"`
	OneRepMax map[string]float64 `json:"oneRepMax"`
}

type logRequest struct {
	CycleID   string `json:"cycleID"`
	Index     int    `json:"index"`
	Date      string `json:"date"`
	AmrapReps int    `json:"amrapReps"`
}

type cycleRequest struct {
	CycleID     string             `json:"cycleID"`
	StartDate   string             `json:"startDate"`
	TrainingMax map[string]float64 `json:"trainingMax"`
}

type errorResponse struct {
	Error string `json:"error"`
}
