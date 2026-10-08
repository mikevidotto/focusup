package workouts

import (
	"fmt"
	"path/filepath"
	"testing"
	"time"
)

func newTestService(t *testing.T) *Service {
	t.Helper()
	return &Service{path: filepath.Join(t.TempDir(), "workouts.json"), cycles: []Cycle{}}
}

var testMaxes = map[string]float64{"squat": 300, "bench": 225, "deadlift": 405, "press": 135}

func setUp(t *testing.T, s *Service) Cycle {
	t.Helper()
	c, err := s.Setup("2026-10-05", testMaxes)
	if err != nil {
		t.Fatal(err)
	}
	return c
}

// day returns 2026-10-05 + n days as a date key.
func day(n int) string {
	return time.Date(2026, 10, 5+n, 0, 0, 0, 0, time.UTC).Format(DateLayout)
}

func logAll(t *testing.T, s *Service, cycleID string, from, to int) {
	t.Helper()
	for i := from; i < to; i++ {
		if _, err := s.CompleteWorkout(cycleID, i, day(i), 5); err != nil {
			t.Fatalf("CompleteWorkout(%d): %v", i, err)
		}
	}
}

func TestSetup_TrainingMaxIsNinetyPercentRounded(t *testing.T) {
	s := newTestService(t)
	c := setUp(t, s)

	want := map[string]float64{"squat": 270, "bench": 205, "deadlift": 365, "press": 120}
	for lift, tm := range want {
		if c.TrainingMax[lift] != tm {
			t.Errorf("%s TM = %v, want %v", lift, c.TrainingMax[lift], tm)
		}
	}

	if _, err := s.Setup("2026-10-05", testMaxes); err != ErrAlreadySetUp {
		t.Fatalf("second Setup err = %v, want ErrAlreadySetUp", err)
	}
	if _, err := newTestService(t).Setup("2026-10-05", map[string]float64{"squat": 300}); err != ErrInvalidInput {
		t.Fatalf("missing lifts err = %v, want ErrInvalidInput", err)
	}
}

func TestCompleteWorkout_EnforcesOrderAndDates(t *testing.T) {
	s := newTestService(t)
	c := setUp(t, s)

	if _, err := s.CompleteWorkout(c.ID, 1, day(0), 5); err != ErrOutOfOrder {
		t.Fatalf("skipping ahead err = %v, want ErrOutOfOrder", err)
	}
	if _, err := s.CompleteWorkout(c.ID, 0, day(2), 8); err != nil {
		t.Fatal(err)
	}
	if _, err := s.CompleteWorkout(c.ID, 1, day(1), 5); err != ErrInvalidDate {
		t.Fatalf("date before previous err = %v, want ErrInvalidDate", err)
	}
	if _, err := s.CompleteWorkout(c.ID, 1, "10/07/2026", 5); err != ErrInvalidDate {
		t.Fatalf("bad format err = %v, want ErrInvalidDate", err)
	}
}

func TestCompleteWorkout_DeloadIgnoresAmrapAndGeneratesNextCycle(t *testing.T) {
	s := newTestService(t)
	c := setUp(t, s)

	logAll(t, s, c.ID, 0, WorkoutsPerCycle-1)
	cycles, err := s.CompleteWorkout(c.ID, WorkoutsPerCycle-1, day(20), 9)
	if err != nil {
		t.Fatal(err)
	}

	if len(cycles) != 2 {
		t.Fatalf("cycles = %d, want 2", len(cycles))
	}
	if reps := cycles[0].Logs[WorkoutsPerCycle-1].AmrapReps; reps != 0 {
		t.Errorf("deload AMRAP reps = %d, want 0", reps)
	}

	next := cycles[1]
	if next.Number != 2 || next.StartDate != day(23) {
		t.Errorf("next cycle = #%d starting %s, want #2 starting %s", next.Number, next.StartDate, day(23))
	}
	want := map[string]float64{"squat": 280, "bench": 210, "deadlift": 375, "press": 125}
	for lift, tm := range want {
		if next.TrainingMax[lift] != tm {
			t.Errorf("next %s TM = %v, want %v", lift, next.TrainingMax[lift], tm)
		}
	}

	if _, err := s.CompleteWorkout(c.ID, 0, day(24), 5); err != ErrNotFound {
		t.Fatalf("logging into finished cycle err = %v, want ErrNotFound", err)
	}
	if _, err := s.CompleteWorkout(next.ID, 0, day(19), 5); err != ErrInvalidDate {
		t.Fatalf("next cycle before last workout err = %v, want ErrInvalidDate", err)
	}
}

func TestUndoLastWorkout_RemovesGeneratedCycle(t *testing.T) {
	s := newTestService(t)
	c := setUp(t, s)

	if _, err := s.UndoLastWorkout(); err != ErrNothingToUndo {
		t.Fatalf("undo with no logs err = %v, want ErrNothingToUndo", err)
	}

	logAll(t, s, c.ID, 0, WorkoutsPerCycle)
	cycles, err := s.UndoLastWorkout()
	if err != nil {
		t.Fatal(err)
	}
	if len(cycles) != 1 || len(cycles[0].Logs) != WorkoutsPerCycle-1 {
		t.Fatalf("after undo: %d cycles, %d logs", len(cycles), len(cycles[0].Logs))
	}
}

func TestUpdateCycleAndLog_Validation(t *testing.T) {
	s := newTestService(t)
	c := setUp(t, s)
	logAll(t, s, c.ID, 0, 3)

	newTM := map[string]float64{"squat": 275, "bench": 200, "deadlift": 360, "press": 115}
	if _, err := s.UpdateCycle(c.ID, day(1), newTM); err != ErrInvalidDate {
		t.Fatalf("start after first log err = %v, want ErrInvalidDate", err)
	}
	updated, err := s.UpdateCycle(c.ID, day(-2), newTM)
	if err != nil {
		t.Fatal(err)
	}
	if updated.TrainingMax["squat"] != 275 || updated.StartDate != day(-2) {
		t.Fatalf("updated cycle = %+v", updated)
	}
	if _, err := s.UpdateCycle(c.ID, day(0), map[string]float64{"squat": 0}); err != ErrInvalidInput {
		t.Fatalf("zero TM err = %v, want ErrInvalidInput", err)
	}

	if _, err := s.UpdateWorkoutLog(c.ID, 1, day(3), 5); err != ErrInvalidDate {
		t.Fatalf("log after next log err = %v, want ErrInvalidDate", err)
	}
	log, err := s.UpdateWorkoutLog(c.ID, 1, day(2), 11)
	if err != nil {
		t.Fatal(err)
	}
	if got := log.Logs[1]; got.Date != day(2) || got.AmrapReps != 11 {
		t.Fatalf("log = %+v", got)
	}
	if _, err := s.UpdateWorkoutLog(c.ID, 5, day(5), 5); err != ErrNotFound {
		t.Fatalf("unlogged index err = %v, want ErrNotFound", err)
	}
}

func TestPersistsAcrossReload(t *testing.T) {
	s := newTestService(t)
	c := setUp(t, s)
	logAll(t, s, c.ID, 0, 2)

	loaded, err := load(s.path)
	if err != nil {
		t.Fatal(err)
	}
	if got := fmt.Sprint(len(loaded), len(loaded[0].Logs), loaded[0].TrainingMax["bench"]); got != "1 2 205" {
		t.Fatalf("reloaded = %s", got)
	}

	if err := s.Reset(); err != nil {
		t.Fatal(err)
	}
	if loaded, _ := load(s.path); len(loaded) != 0 {
		t.Fatalf("after reset %d cycles", len(loaded))
	}
}
