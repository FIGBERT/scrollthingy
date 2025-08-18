package main

import (
	"cmp"
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"time"

	"github.com/fcjr/scroll-together/server/internal/server"
	"github.com/joho/godotenv"
	"github.com/lmittmann/tint"
)

func main() {
	ctx := context.Background()

	err := godotenv.Load()
	if err != nil {
		slog.Error(fmt.Sprintf("Error loading .env file: %s, skipping...", err))
	}

	if err := run(ctx, os.Args, os.Getenv, os.Stdin, os.Stdout, os.Stderr); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, args []string, getenv func(string) string, stdin io.Reader, stdout, stderr io.Writer) error {
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	addr := cmp.Or(getenv("ADDR"), ":8080")
	debug := cmp.Or(getenv("DEBUG"), "false") == "true"

	est, err := time.LoadLocation("America/New_York")
	if err != nil {
		return err
	}

	logLevel := slog.LevelInfo
	if debug {
		logLevel = slog.LevelDebug
	}

	logger := slog.New(tint.NewHandler(os.Stderr, &tint.Options{
		AddSource:  true,
		Level:      logLevel,
		TimeFormat: "2006-01-02 03:04:05 PM MST",
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey {
				if t, ok := a.Value.Any().(time.Time); ok {
					// Convert the time to EST
					return slog.Time(slog.TimeKey, t.In(est))
				}
			}
			return a
		},
	}))

	var options = []func(*server.Server) error{
		server.WithLogger(logger),
	}

	s, err := server.New(server.NewParams{
		Logger: logger,
	},
		options...)

	if err != nil {
		logger.Error("could not create server", "error", err)
		return err
	}
	return s.ListenAndServe(ctx, addr)
}
