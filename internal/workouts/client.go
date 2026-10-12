package workouts

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"
)

var ErrServerUnreachable = errors.New("can't reach the workout server")

// Client is a Backend that keeps the program on the gym server (see
// NewHandler), so workouts logged from the phone show up on the desktop
// and the other way round. The last cycle list it fetched is cached in
// the local data file: when the server is down ListCycles shows that copy,
// but every change fails rather than letting the two copies drift apart.
type Client struct {
	baseURL string
	http    *http.Client

	mu        sync.Mutex
	cachePath string
	offline   bool
}

func NewClient(baseURL string) (*Client, error) {
	path, err := datafilepath()
	if err != nil {
		return nil, err
	}
	return newClientAt(baseURL, path), nil
}

func newClientAt(baseURL, cachePath string) *Client {
	return &Client{
		baseURL:   strings.TrimRight(baseURL, "/"),
		http:      &http.Client{Timeout: 10 * time.Second},
		cachePath: cachePath,
	}
}

// Offline reports whether the last ListCycles had to fall back to the cache.
func (c *Client) Offline() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.offline
}

func (c *Client) ListCycles() []Cycle {
	var cycles []Cycle
	err := c.call(http.MethodGet, "/api/cycles", nil, &cycles)

	c.mu.Lock()
	defer c.mu.Unlock()

	if err != nil {
		c.offline = true
		cached, loadErr := load(c.cachePath)
		if loadErr != nil {
			return []Cycle{}
		}
		return cached
	}

	c.offline = false
	save(c.cachePath, cycles)
	return cycles
}

func (c *Client) Setup(startDate string, oneRepMax map[string]float64) (Cycle, error) {
	var out Cycle
	err := c.write("/api/setup", setupRequest{StartDate: startDate, OneRepMax: oneRepMax}, &out)
	return out, err
}

func (c *Client) CompleteWorkout(cycleID string, index int, date string, amrapReps int) ([]Cycle, error) {
	var out []Cycle
	err := c.write("/api/complete", logRequest{CycleID: cycleID, Index: index, Date: date, AmrapReps: amrapReps}, &out)
	return out, err
}

func (c *Client) UpdateCycle(cycleID, startDate string, trainingMax map[string]float64) (Cycle, error) {
	var out Cycle
	err := c.write("/api/cycle", cycleRequest{CycleID: cycleID, StartDate: startDate, TrainingMax: trainingMax}, &out)
	return out, err
}

func (c *Client) UpdateWorkoutLog(cycleID string, index int, date string, amrapReps int) (Cycle, error) {
	var out Cycle
	err := c.write("/api/log", logRequest{CycleID: cycleID, Index: index, Date: date, AmrapReps: amrapReps}, &out)
	return out, err
}

func (c *Client) UndoLastWorkout() ([]Cycle, error) {
	var out []Cycle
	err := c.write("/api/undo", nil, &out)
	return out, err
}

func (c *Client) Reset() error {
	return c.write("/api/reset", nil, nil)
}

// write sends a change, then refreshes the cache so it never lags behind
// what the desktop last showed.
func (c *Client) write(path string, body, out any) error {
	if err := c.call(http.MethodPost, path, body, out); err != nil {
		return err
	}
	c.ListCycles()
	return nil
}

func (c *Client) call(method, path string, body, out any) error {
	var payload bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&payload).Encode(body); err != nil {
			return err
		}
	}

	req, err := http.NewRequest(method, c.baseURL+path, &payload)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("%w at %s", ErrServerUnreachable, c.baseURL)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		var e errorResponse
		if json.NewDecoder(resp.Body).Decode(&e) != nil || e.Error == "" {
			e.Error = resp.Status
		}
		return errors.New(e.Error)
	}

	if out == nil || resp.StatusCode == http.StatusNoContent {
		return nil
	}
	return json.NewDecoder(resp.Body).Decode(out)
}
