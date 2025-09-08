package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"time"

	"github.com/joho/godotenv"
	"github.com/lmittmann/tint"

	"github.com/fcjr/scroll-together/server/internal/server"
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
		logger.Error("could not load .env (api keys likely empty)")
	}

	s, err := server.New(logger)
	if err != nil {
		logger.Error("could not create server", "err", err)
		os.Exit(1)
	}
	defer s.Cleanup()

	err = s.ListenAndServe(ctx, ":8080")
	if err != nil {
		os.Exit(1)
	}
}
