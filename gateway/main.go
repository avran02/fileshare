package main

import (
	"log/slog"
	"os"

	"github.com/avran02/fileshare/gateway/internal/app"
)

func main() {
	app := app.New()
	err := app.RunServer()
	if err != nil {
		slog.Error("Failed to run server:\n" + err.Error())
		os.Exit(1)
	}
}
