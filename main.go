package main

import (
	"context"
	"embed"
	"io/fs"
	"log"

	"garq/internal/api"
	"garq/internal/db"
	"garq/internal/worker"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed cmd/wails/frontend/*
var assets embed.FS

func main() {
	appFS, err := fs.Sub(assets, "cmd/wails/frontend")
	if err != nil {
		log.Fatal(err)
	}

	conn, err := db.InitDB("garq.db")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	worker.StartWorkerPool(4, conn)
	bindAPI := &api.API{DB: conn}

	err = wails.Run(&options.App{
		Title:     "Garq",
		Width:     1280,
		Height:    800,
		OnStartup: func(ctx context.Context) {},
		AssetServer: &assetserver.Options{
			Assets: appFS,
		},
		Bind: []interface{}{
			bindAPI,
		},
	})
	if err != nil {
		log.Fatal(err)
	}
}
