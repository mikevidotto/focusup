package main

import (
	"context"
	"log"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"

	appservice "focusup/internal/app"
	"focusup/internal/calendar"
	"focusup/internal/habits"
	"focusup/internal/jobs"
	"focusup/internal/journal"
	"focusup/internal/notifier"
	"focusup/internal/settings"
	"focusup/internal/tasks"
	"focusup/internal/workouts"
)

// reminderDueEvent is the Wails runtime event name the frontend listens on
// (via EventsOn) to show an in-app popup when a reminder fires.
const reminderDueEvent = "calendar:reminder-due"

type App struct {
	ctx         context.Context
	infoService *appservice.InfoService
	tasks       *tasks.Service
	habits      *habits.Service
	journal     *journal.Service
	calendar    *calendar.Service
	notifier    *notifier.Notifier
	settings    *settings.Service
	workouts    *workouts.Service
	jobs        *jobs.Service
}

func NewApp() *App {
	taskService, err := tasks.NewService()
	if err != nil {
		log.Fatalf("failed to initialize task storage: %v", err)
	}

	habitService, err := habits.NewService()
	if err != nil {
		log.Fatalf("failed to initialize habit storage: %v", err)
	}

	journalService, err := journal.NewService()
	if err != nil {
		log.Fatalf("failed to initialize journal storage: %v", err)
	}

	calendarService, err := calendar.NewService()
	if err != nil {
		log.Fatalf("failed to initialize calendar storage: %v", err)
	}

	workoutService, err := workouts.NewService()
	if err != nil {
		log.Fatalf("failed to initialize workout storage: %v", err)
	}

	settingsService, err := settings.NewService()
	if err != nil {
		log.Fatalf("failed to initialize settings storage: %v", err)
	}

	app := &App{
		infoService: appservice.NewInfoService(),
		tasks:       taskService,
		habits:      habitService,
		journal:     journalService,
		calendar:    calendarService,
		settings:    settingsService,
		workouts:    workoutService,
	}

	app.jobs = jobs.NewService(settingsService.Get().JobSearchDir)

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

func (a *App) GetSettings() settings.Settings {
	return a.settings.Get()
}

// SetTheme persists the UI theme ("dark" or "light").
func (a *App) SetTheme(theme string) (settings.Settings, error) {
	return a.settings.SetTheme(theme)
}

// MarkWeeklyReviewed records that the GTD weekly review was just finished.
func (a *App) MarkWeeklyReviewed() (settings.Settings, error) {
	return a.settings.MarkReviewed()
}

func (a *App) ListTasks() []tasks.Task {
	return a.tasks.List()
}

// AddTask captures a task into the GTD inbox.
func (a *App) AddTask(title string, contexts []string) (tasks.Task, error) {
	return a.tasks.Add(title, contexts)
}

func (a *App) AddProjectTask(projectID, title string, contexts []string) (tasks.Task, error) {
	return a.tasks.AddProjectTask(projectID, title, contexts)
}

func (a *App) ToggleTask(id string) (tasks.Task, error) {
	return a.tasks.Toggle(id)
}

// MoveTask puts a task in another GTD list ("inbox", "next" or "someday").
func (a *App) MoveTask(id, list string) (tasks.Task, error) {
	return a.tasks.Move(id, list)
}

func (a *App) RenameTask(id, title string) (tasks.Task, error) {
	return a.tasks.Rename(id, title)
}

func (a *App) SetTaskContexts(id string, contexts []string) (tasks.Task, error) {
	return a.tasks.SetContexts(id, contexts)
}

// SetTaskProject links a task to a project; an empty projectID unlinks it.
func (a *App) SetTaskProject(id, projectID string) (tasks.Task, error) {
	return a.tasks.SetProject(id, projectID)
}

func (a *App) DeleteTask(id string) error {
	return a.tasks.Delete(id)
}

func (a *App) ListProjects() []tasks.Project {
	return a.tasks.ListProjects()
}

func (a *App) AddProject(title string) (tasks.Project, error) {
	return a.tasks.AddProject(title)
}

func (a *App) RenameProject(id, title string) (tasks.Project, error) {
	return a.tasks.RenameProject(id, title)
}

func (a *App) ToggleProject(id string) (tasks.Project, error) {
	return a.tasks.ToggleProject(id)
}

func (a *App) DeleteProject(id string) error {
	return a.tasks.DeleteProject(id)
}

func (a *App) ListHabits() []habits.Habit {
	return a.habits.List()
}

func (a *App) AddHabit(name string) (habits.Habit, error) {
	return a.habits.Add(name)
}

func (a *App) RenameHabit(id, name string) (habits.Habit, error) {
	return a.habits.Rename(id, name)
}

// ToggleHabitCompletion flips whether the habit was done on date (YYYY-MM-DD).
func (a *App) ToggleHabitCompletion(id, date string) (habits.Habit, error) {
	return a.habits.ToggleCompletion(id, date)
}

func (a *App) DeleteHabit(id string) error {
	return a.habits.Delete(id)
}

func (a *App) ListJournalEntries() []journal.Entry {
	return a.journal.List()
}

// SaveJournalEntry creates or replaces the entry for date (YYYY-MM-DD); an
// entry with nothing written in it is removed instead.
func (a *App) SaveJournalEntry(date string, prompts []journal.PromptAnswer, body string) (journal.Entry, error) {
	return a.journal.Save(date, prompts, body)
}

func (a *App) DeleteJournalEntry(date string) error {
	return a.journal.Delete(date)
}

func (a *App) ListWorkoutCycles() []workouts.Cycle {
	return a.workouts.ListCycles()
}

// SetupWorkouts starts the 5/3/1 program from the lifter's true 1RMs.
func (a *App) SetupWorkouts(startDate string, oneRepMax map[string]float64) (workouts.Cycle, error) {
	return a.workouts.Setup(startDate, oneRepMax)
}

// CompleteWorkout logs the next workout; it returns every cycle because
// finishing a cycle generates the next one.
func (a *App) CompleteWorkout(cycleID string, index int, date string, amrapReps int) ([]workouts.Cycle, error) {
	return a.workouts.CompleteWorkout(cycleID, index, date, amrapReps)
}

func (a *App) UpdateWorkoutCycle(cycleID, startDate string, trainingMax map[string]float64) (workouts.Cycle, error) {
	return a.workouts.UpdateCycle(cycleID, startDate, trainingMax)
}

func (a *App) UpdateWorkoutLog(cycleID string, index int, date string, amrapReps int) (workouts.Cycle, error) {
	return a.workouts.UpdateWorkoutLog(cycleID, index, date, amrapReps)
}

func (a *App) UndoLastWorkout() ([]workouts.Cycle, error) {
	return a.workouts.UndoLastWorkout()
}

func (a *App) ResetWorkouts() error {
	return a.workouts.Reset()
}

func (a *App) ListEvents() []calendar.Event {
	return a.calendar.List()
}

func (a *App) ListCalendarOccurrences(rangeStart, rangeEnd time.Time) []calendar.OccurrenceView {
	return a.calendar.ListOccurrences(rangeStart, rangeEnd)
}

func (a *App) AddEvent(input calendar.EventInput) (calendar.Event, error) {
	return a.calendar.Add(input)
}

func (a *App) UpdateEvent(id string, input calendar.EventInput) (calendar.Event, error) {
	return a.calendar.Update(id, input)
}

// SkipEventOccurrence removes one occurrence of a recurring event, keyed by
// the occurrence's originalStart.
func (a *App) SkipEventOccurrence(eventID string, originalDate time.Time) (calendar.Event, error) {
	return a.calendar.SkipOccurrence(eventID, originalDate)
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

// JobsInfo describes the ai-job-search repo the Jobs tab is pointed at.
type JobsInfo struct {
	Dir string `json:"dir"`
}

func (a *App) GetJobsInfo() JobsInfo {
	return JobsInfo{Dir: a.jobs.Dir()}
}

// SetJobSearchDir points the Jobs tab at another ai-job-search checkout.
func (a *App) SetJobSearchDir(dir string) (JobsInfo, error) {
	saved, err := a.settings.SetJobSearchDir(dir)
	if err != nil {
		return JobsInfo{}, err
	}

	a.jobs.SetDir(saved.JobSearchDir)
	return a.GetJobsInfo(), nil
}

func (a *App) ListJobQueue() ([]jobs.QueueItem, error) {
	return a.jobs.ListQueue()
}

func (a *App) ListJobApplications() ([]jobs.Application, error) {
	return a.jobs.ListApplications()
}

// SetJobQueueOutcome ticks an apply-queue line, e.g. "skipped - <reason>".
func (a *App) SetJobQueueOutcome(number int, outcome string) error {
	return a.jobs.SetQueueOutcome(number, outcome)
}

// MarkJobApplied marks a tracker row (by index) applied today and ticks the
// matching queue line.
func (a *App) MarkJobApplied(row int) error {
	return a.jobs.MarkApplied(row, time.Now().Format("2006-01-02"))
}
