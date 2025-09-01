package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"time"

	"github.com/lmittmann/tint"

	"github.com/figbert/scroll-server/internal/server"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	logger := slog.New(tint.NewHandler(os.Stderr, &tint.Options{
		AddSource:  true,
		TimeFormat: time.DateTime,
	}))
	run(logger)

	s, err := server.New(server.WithLogger(logger))
	if err != nil {
		logger.Error("could not create server", "error", err)
	}

	if err = s.ListenAndServe(ctx, ":8080"); err != nil {
		os.Exit(1)
	}
}
