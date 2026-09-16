package main

import (
	"context"
	"log"

	appservice "focusup/internal/app"
	"focusup/internal/tasks"
)

type App struct {
	ctx         context.Context
	infoService *appservice.InfoService
	tasks       *tasks.Service
}

func NewApp() *App {
	taskService, err := tasks.NewService()
	if err != nil {
		log.Fatalf("failed to initialize task storage: %v", err)
	}

	return &App{
		infoService: appservice.NewInfoService(),
		tasks:       taskService,
	}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
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

func (a *App) AddTask(title string) (tasks.Task, error) {
	return a.tasks.Add(title)
}

func (a *App) ToggleTask(id string) (tasks.Task, error) {
	return a.tasks.Toggle(id)
}

func (a *App) DeleteTask(id string) error {
	return a.tasks.Delete(id)
}
