package main

import (
	"embed"
	"log"

	"github.com/wailsapp/wails/v3/pkg/application"

	"video-extractor/internal/app"
	"video-extractor/internal/updater"
)

var version = "dev"

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	appService := app.NewService(version)

	appInstance := application.New(application.Options{
		Name:        "Extral",
		Description: "视频抽帧桌面客户端",
		Services: []application.Service{
			application.NewService(appService),
			application.NewService(updater.NewService(version, "wangjin/video-extractor")),
		},
		Assets: application.AssetOptions{
			Handler: application.BundledAssetFileServer(assets),
		},
		Mac: application.MacOptions{
			ApplicationShouldTerminateAfterLastWindowClosed: true,
		},
	})

	appInstance.Window.NewWithOptions(application.WebviewWindowOptions{
		Title:           "Extral",
		Width:           960,
		Height:          720,
		DevToolsEnabled: true,
	})

	if err := appInstance.Run(); err != nil {
		log.Fatal(err)
	}
}
