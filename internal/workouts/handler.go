package workouts

import (
	"encoding/json"
	"errors"
	"net/http"
)

// NewHandler serves s as a JSON API under /api/. Every route maps to one
// Service method; Client is its counterpart.
//
//	GET  /api/cycles    ListCycles
//	POST /api/setup     Setup
//	POST /api/complete  CompleteWorkout
//	POST /api/cycle     UpdateCycle
//	POST /api/log       UpdateWorkoutLog
//	POST /api/undo      UndoLastWorkout
//	POST /api/reset     Reset
func NewHandler(s *Service) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/cycles", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, s.ListCycles())
	})

	mux.HandleFunc("POST /api/setup", func(w http.ResponseWriter, r *http.Request) {
		var req setupRequest
		if !decode(w, r, &req) {
			return
		}
		result, err := s.Setup(req.StartDate, req.OneRepMax)
		respond(w, result, err)
	})

	mux.HandleFunc("POST /api/complete", func(w http.ResponseWriter, r *http.Request) {
		var req logRequest
		if !decode(w, r, &req) {
			return
		}
		result, err := s.CompleteWorkout(req.CycleID, req.Index, req.Date, req.AmrapReps)
		respond(w, result, err)
	})

	mux.HandleFunc("POST /api/cycle", func(w http.ResponseWriter, r *http.Request) {
		var req cycleRequest
		if !decode(w, r, &req) {
			return
		}
		result, err := s.UpdateCycle(req.CycleID, req.StartDate, req.TrainingMax)
		respond(w, result, err)
	})

	mux.HandleFunc("POST /api/log", func(w http.ResponseWriter, r *http.Request) {
		var req logRequest
		if !decode(w, r, &req) {
			return
		}
		result, err := s.UpdateWorkoutLog(req.CycleID, req.Index, req.Date, req.AmrapReps)
		respond(w, result, err)
	})

	mux.HandleFunc("POST /api/undo", func(w http.ResponseWriter, r *http.Request) {
		result, err := s.UndoLastWorkout()
		respond(w, result, err)
	})

	mux.HandleFunc("POST /api/reset", func(w http.ResponseWriter, r *http.Request) {
		if err := s.Reset(); err != nil {
			writeError(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})

	return mux
}

func decode(w http.ResponseWriter, r *http.Request, into any) bool {
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(into); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid request body"})
		return false
	}
	return true
}

func respond[T any](w http.ResponseWriter, result T, err error) {
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

// userErrors are the program's validation errors: the request was wrong,
// not the server.
var userErrors = []error{ErrNotFound, ErrOutOfOrder, ErrInvalidDate, ErrInvalidInput, ErrAlreadySetUp, ErrNothingToUndo}

func writeError(w http.ResponseWriter, err error) {
	status := http.StatusInternalServerError
	for _, e := range userErrors {
		if errors.Is(err, e) {
			status = http.StatusBadRequest
			break
		}
	}
	writeJSON(w, status, errorResponse{Error: err.Error()})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
