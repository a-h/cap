package main

import (
	"context"
	"fmt"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App is the Wails application backend.
type App struct {
	ctx     context.Context
	handler *switchableHandler
}

// SelectFolder opens a native directory picker and configures the cap model root.
// Returns the selected path, or an empty string if the dialog was cancelled.
func (a *App) SelectFolder() string {
	dir, err := runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Select cap root folder",
	})
	if err != nil || dir == "" {
		return ""
	}
	a.handler.SetRoot(dir)
	runtime.WindowSetTitle(a.ctx, fmt.Sprintf("cap - %s", dir))
	return dir
}
