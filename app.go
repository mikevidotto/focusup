package main

import (
	"context"

	appservice "focusup/internal/app"
)

type App struct {
	ctx         context.Context
	infoService *appservice.InfoService
}

func NewApp() *App {
	return &App{
		infoService: appservice.NewInfoService(),
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
