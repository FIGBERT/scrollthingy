package server

import (
	"context"
	"log/slog"
	"net"
	"net/http"
	"time"

	lksdk "github.com/livekit/server-sdk-go/v2"

	"github.com/fcjr/scroll-together/server/internal/camera"
	"github.com/fcjr/scroll-together/server/internal/middleware"
)

type Server struct {
	rig    *camera.Rig
	logger *slog.Logger
	room   *lksdk.Room
}

func New(rig *camera.Rig, logger *slog.Logger) (*Server, error) {
	room, err := join_room()
	if err != nil {
		logger.Error("unable to connect to livekit", "room", ROOM_NAME, "err", err)
	}

	s := &Server{
		rig:    rig,
		logger: logger,
		room:   room,
	}
	return s, nil
}

func (s *Server) ListenAndServe(ctx context.Context, addr string) error {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /token", s.token())
	s.publish_camera()
	s.room.RegisterTextStreamHandler("scroll-updates", func(reader *lksdk.TextStreamReader, participant string) {
		s.logger.Info("receiver msg", "text", reader.ReadAll(), "participant", participant, "topic", "scroll-updates")
	})
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
