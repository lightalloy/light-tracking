package main

import (
	"embed"
	"log"

	"light-tracking/internal/app"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	// Create an instance of the app structure
	appInstance, err := app.NewApp()
	if err != nil {
		log.Fatal("Failed to create app:", err)
	}
	defer appInstance.Close()

	// Create application with options
	err = wails.Run(&options.App{
		Title:  "Light Tracking",
		Width:  800,
		Height: 600,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 255, G: 255, B: 255, A: 1},
		OnStartup:        appInstance.Startup,
		Bind: []interface{}{
			appInstance,
		},
	})

	if err != nil {
		log.Fatal("Error:", err)
	}
}
