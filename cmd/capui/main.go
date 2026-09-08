package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/menu"
	"github.com/wailsapp/wails/v2/pkg/menu/keys"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

func main() {
	sw := &switchableHandler{}
	app := &App{handler: sw}

	appMenu := menu.NewMenu()
	fileMenu := appMenu.AddSubmenu("File")
	fileMenu.AddText("Open Folder...", keys.CmdOrCtrl("o"), func(_ *menu.CallbackData) {
		dir, err := runtime.OpenDirectoryDialog(app.ctx, runtime.OpenDialogOptions{
			Title: "Select cap root folder",
		})
		if err != nil || dir == "" {
			return
		}
		sw.SetRoot(dir)
		runtime.WindowSetTitle(app.ctx, "cap - "+dir)
		runtime.WindowExecJS(app.ctx, "window.location.href = '/'")
	})

	err := wails.Run(&options.App{
		Title:  "cap",
		Width:  1280,
		Height: 800,
		Menu:   appMenu,
		AssetServer: &assetserver.Options{
			Handler: sw,
		},
		OnStartup: func(ctx context.Context) { app.ctx = ctx },
		Bind:      []any{app},
	})
	if err != nil {
		slog.Error("failed to run application", slog.Any("err", err))
		os.Exit(1)
	}
}
