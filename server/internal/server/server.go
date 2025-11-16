package server

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"time"

	lksdk "github.com/livekit/server-sdk-go/v2"

	"github.com/fcjr/scroll-together/server/internal/camera"
	"github.com/fcjr/scroll-together/server/internal/middleware"
)

type state struct {
	current string
	users   []string
}

type Server struct {
	rig    *camera.Rig
	logger *slog.Logger
	room   *lksdk.Room

	state *state
	ports *ports
}

func New(logger *slog.Logger) (*Server, error) {
	if logger == nil {
		return nil, fmt.Errorf("a logger is required to initialize the server")
	}

	s := &Server{logger: logger, state: &state{users: make([]string, 0)}}

	err := s.setupCamera()
	if err != nil {
		logger.Error("unable to create camera rig")
		return nil, err
	}

	err = s.joinRoom()
	if err != nil {
		logger.Error("unable to connect to livekit", "room", ROOM_NAME)
		return nil, err
	}

	err, port := s.setupGPIO()
	if err != nil {
		logger.Error("unable to bind to gpio port(s)", "id", *port)
		return nil, err
	}

	return s, nil
}

func (s *Server) Cleanup() {
	if s.rig != nil {
		s.rig.Reader.Close()
		s.rig.Track.Close()
	}
	s.CleanupGPIO()
}

func (s *Server) ListenAndServe(ctx context.Context, addr string) error {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /token", s.token())
	mux.HandleFunc("GET /", s.index())
	mux.HandleFunc("GET /client.mjs", s.script())

	s.publishCamera()
	s.room.RegisterTextStreamHandler("scroll-updates", s.handleScroll)
	defer s.room.Disconnect()

	// apply middlewares
	handler := middleware.Chain(mux,
		middleware.WithPanicRecovery(s.logger),
		middleware.WithRequestResponseLogging(s.logger),
	)
	httpServer := &http.Server{
		Addr:         addr,
		Handler:      handler,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		BaseContext: func(listener net.Listener) context.Context {
			return ctx
		},
	}

	// start server
	s.logger.Info("server listening", "addr", addr)

	errCh := make(chan error, 1)
	go func() {
		errCh <- httpServer.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		s.logger.Info("server shutting down")
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		return httpServer.Shutdown(ctx)
	case err := <-errCh:
		s.logger.Error("server listen error, shuting down", "err", err)
		return err
	}
}
