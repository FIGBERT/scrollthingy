package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"time"

	"github.com/joho/godotenv"
	"github.com/lmittmann/tint"

	"github.com/figbert/scroll-server/internal/camera"
	"github.com/figbert/scroll-server/internal/server"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	logger := slog.New(tint.NewHandler(os.Stderr, &tint.Options{
		AddSource:  true,
		TimeFormat: time.DateTime,
	}))

	err := godotenv.Load()
	if err != nil {
		logger.Error("could not load .env (api keys will be empty)")
	}

	rig, err := camera.Setup()
	if err != nil {
		logger.Error("could not creat camera rig", "error", err)
		os.Exit(1)
	}
	defer rig.Reader.Close()
	defer rig.Track.Close()

	s, err := server.New(rig, logger)
	if err != nil {
		logger.Error("could not create server", "error", err)
		os.Exit(1)
	}

	err = s.ListenAndServe(ctx, ":8080")
	if err != nil {
		os.Exit(1)
	}
}
