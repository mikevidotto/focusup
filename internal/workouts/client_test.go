package workouts

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
)

// newTestClient runs a Service behind NewHandler and points a Client at it.
func newTestClient(t *testing.T) (*Client, *Service, *httptest.Server) {
	t.Helper()
	s := newTestService(t)
	srv := httptest.NewServer(NewHandler(s))
	t.Cleanup(srv.Close)
	return newClientAt(srv.URL+"/", filepath.Join(t.TempDir(), "cache.json")), s, srv
}

func TestClient_RoundTripsThroughHandler(t *testing.T) {
	c, s, _ := newTestClient(t)

	cycle, err := c.Setup("2026-10-05", testMaxes)
	if err != nil {
		t.Fatal(err)
	}
	if cycle.TrainingMax["squat"] != 270 {
		t.Fatalf("squat TM = %v, want 270", cycle.TrainingMax["squat"])
	}

	cycles, err := c.CompleteWorkout(cycle.ID, 0, day(0), 8)
	if err != nil {
		t.Fatal(err)
	}
	if got := cycles[0].Logs[0].AmrapReps; got != 8 {
		t.Fatalf("logged reps = %d, want 8", got)
	}

	// The server's copy is the one that changed.
	if got := len(s.ListCycles()[0].Logs); got != 1 {
		t.Fatalf("server logs = %d, want 1", got)
	}

	updated, err := c.UpdateWorkoutLog(cycle.ID, 0, day(0), 9)
	if err != nil {
		t.Fatal(err)
	}
	if updated.Logs[0].AmrapReps != 9 {
		t.Fatalf("updated reps = %d, want 9", updated.Logs[0].AmrapReps)
	}

	tm := map[string]float64{"squat": 275, "bench": 205, "deadlift": 365, "press": 120}
	if _, err := c.UpdateCycle(cycle.ID, "2026-10-05", tm); err != nil {
		t.Fatal(err)
	}

	cycles, err = c.UndoLastWorkout()
	if err != nil {
		t.Fatal(err)
	}
	if len(cycles[0].Logs) != 0 {
		t.Fatalf("logs after undo = %d, want 0", len(cycles[0].Logs))
	}

	if err := c.Reset(); err != nil {
		t.Fatal(err)
	}
	if got := len(c.ListCycles()); got != 0 {
		t.Fatalf("cycles after reset = %d, want 0", got)
	}
}

func TestClient_PassesValidationErrorsThrough(t *testing.T) {
	c, _, srv := newTestClient(t)
	cycle, err := c.Setup("2026-10-05", testMaxes)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := c.CompleteWorkout(cycle.ID, 3, day(0), 5); err == nil || err.Error() != ErrOutOfOrder.Error() {
		t.Fatalf("out of order err = %v, want %v", err, ErrOutOfOrder)
	}

	resp, err := http.Post(srv.URL+"/api/complete", "application/json", strings.NewReader(`{"cycleID":"nope","index":0,"date":"2026-10-05"}`))
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", resp.StatusCode)
	}
}

func TestClient_FallsBackToCacheWhenOffline(t *testing.T) {
	c, _, srv := newTestClient(t)
	cycle, err := c.Setup("2026-10-05", testMaxes)
	if err != nil {
		t.Fatal(err)
	}
	if c.Offline() {
		t.Fatal("Offline() = true while the server is up")
	}

	srv.Close()

	cached := c.ListCycles()
	if len(cached) != 1 || cached[0].ID != cycle.ID {
		t.Fatalf("cached cycles = %+v, want the set-up cycle", cached)
	}
	if !c.Offline() {
		t.Fatal("Offline() = false after the server went away")
	}

	if _, err := c.CompleteWorkout(cycle.ID, 0, day(0), 5); err == nil || !strings.Contains(err.Error(), ErrServerUnreachable.Error()) {
		t.Fatalf("write while offline err = %v, want ErrServerUnreachable", err)
	}
}
