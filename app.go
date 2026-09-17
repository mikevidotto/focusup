package main

import (
	"context"
	"log"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"

	appservice "focusup/internal/app"
	"focusup/internal/calendar"
	"focusup/internal/notifier"
	"focusup/internal/tasks"
)

// reminderDueEvent is the Wails runtime event name the frontend listens on
// (via EventsOn) to show an in-app popup when a reminder fires.
const reminderDueEvent = "calendar:reminder-due"

type App struct {
	ctx         context.Context
	infoService *appservice.InfoService
	tasks       *tasks.Service
	calendar    *calendar.Service
	notifier    *notifier.Notifier
}

func NewApp() *App {
	taskService, err := tasks.NewService()
	if err != nil {
		log.Fatalf("failed to initialize task storage: %v", err)
	}

	calendarService, err := calendar.NewService()
	if err != nil {
		log.Fatalf("failed to initialize calendar storage: %v", err)
	}

	app := &App{
		infoService: appservice.NewInfoService(),
		tasks:       taskService,
		calendar:    calendarService,
	}

	app.notifier = notifier.New(calendarService, func(reminder calendar.DueReminder) {
		runtime.EventsEmit(app.ctx, reminderDueEvent, reminder)
	})

	return app
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	go a.notifier.Run(ctx)
}

// GetAppInfo is an example of the frontend calling the Go backend.
// Keep widget-specific methods out of App as the project grows.
// Give each major feature/service its own package instead.
func (a *App) GetAppInfo() appservice.Info {
	return a.infoService.GetInfo()
}

func (a *App) ListTasks() []tasks.Task {
	return a.tasks.List()
}

func (a *App) AddTask(title, priority string) (tasks.Task, error) {
	return a.tasks.Add(title, priority)
}

func (a *App) ToggleTask(id string) (tasks.Task, error) {
	return a.tasks.Toggle(id)
}

func (a *App) DeleteTask(id string) error {
	return a.tasks.Delete(id)
}

func (a *App) ListEvents() []calendar.Event {
	return a.calendar.List()
}

func (a *App) ListCalendarOccurrences(rangeStart, rangeEnd time.Time) []calendar.OccurrenceView {
	return a.calendar.ListOccurrences(rangeStart, rangeEnd)
}

func (a *App) AddEvent(title, description, location string, start, end time.Time, allDay bool, recurrence *calendar.RecurrenceRule, important bool) (calendar.Event, error) {
	return a.calendar.Add(title, description, location, start, end, allDay, recurrence, important)
}

func (a *App) AddReminder(eventID string, leadTimeSeconds int) (calendar.Event, error) {
	return a.calendar.AddReminder(eventID, time.Duration(leadTimeSeconds)*time.Second)
}

func (a *App) DeleteEvent(id string) error {
	return a.calendar.Delete(id)
}

func (a *App) ToggleEventCompletion(eventID string, occurrenceDate time.Time) (calendar.Event, error) {
	return a.calendar.ToggleCompletion(eventID, occurrenceDate)
}

func (a *App) GetDueReminders() []calendar.DueReminder {
	return a.calendar.DueReminders(time.Now())
}
