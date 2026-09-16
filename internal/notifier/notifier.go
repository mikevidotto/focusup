// Package notifier polls a calendar.Service for due reminders and fires
// each one exactly once: an OS toast (via internal/notify) plus a
// caller-supplied hook the app uses to notify the frontend for an in-app
// popup.
package notifier

import (
	"context"
	"log"
	"sync"
	"time"

	"focusup/internal/calendar"
	"focusup/internal/notify"
)

const pollInterval = 30 * time.Second

// forgetAfter bounds how long a fired reminder's dedup entry is kept once
// its occurrence has passed, so a long-running app doesn't accumulate an
// ever-growing map.
const forgetAfter = 24 * time.Hour

// DueHandler is invoked once for every reminder that newly becomes due, in
// addition to the OS toast this package always pushes.
type DueHandler func(calendar.DueReminder)

type Notifier struct {
	service *calendar.Service
	onDue   DueHandler

	mu    sync.Mutex
	fired map[string]time.Time // dedup key -> occurrence start, for pruning
}

func New(service *calendar.Service, onDue DueHandler) *Notifier {
	return &Notifier{
		service: service,
		onDue:   onDue,
		fired:   make(map[string]time.Time),
	}
}

// Run polls for due reminders every pollInterval until ctx is canceled.
func (n *Notifier) Run(ctx context.Context) {
	n.checkOnce()

	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			n.checkOnce()
		}
	}
}

func dedupKey(r calendar.DueReminder) string {
	return r.EventID + "|" + r.ReminderID + "|" + r.OccurrenceStart.String()
}

func (n *Notifier) checkOnce() {
	due := n.service.DueReminders(time.Now())

	n.mu.Lock()
	defer n.mu.Unlock()

	for key, occurrenceStart := range n.fired {
		if time.Since(occurrenceStart) > forgetAfter {
			delete(n.fired, key)
		}
	}

	for _, reminder := range due {
		key := dedupKey(reminder)
		if _, alreadyFired := n.fired[key]; alreadyFired {
			continue
		}
		n.fired[key] = reminder.OccurrenceStart

		body := reminder.OccurrenceStart.Format("Mon, Jan 2 · 3:04 PM")
		if err := notify.Push(reminder.EventTitle, body); err != nil {
			log.Printf("notifier: failed to push OS toast: %v", err)
		}

		if n.onDue != nil {
			n.onDue(reminder)
		}
	}
}
