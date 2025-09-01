package server

import (
	"context"
	"log/slog"
	"net"
	"net/http"
	"time"

	"github.com/figbert/scroll-server/internal/middleware"
)

type Server struct {
	logger *slog.Logger
}

func New(opts ...func(*Server) error) (*Server, error) {
	s := &Server{}
	for _, opt := range opts {
		if err := opt(s); err != nil {
			return nil, err
		}
	}
	return s, nil
}

func (s *Server) ListenAndServe(ctx context.Context, addr string) error {
	mux := http.NewServeMux()

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

func (s *Server) redirectTo(path string) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		http.Redirect(res, req, path, http.StatusTemporaryRedirect)
	}
}

func WithLogger(logger *slog.Logger) func(*Server) error {
	return func(s *Server) error {
		s.logger = logger
		return nil
	}
}
