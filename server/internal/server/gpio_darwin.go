package server

import (
	"strconv"

	lksdk "github.com/livekit/server-sdk-go/v2"
)

type ports struct{}

func (s *Server) setupGPIO() (error, *int) { return nil, nil }
func (s *Server) CleanupGPIO()             {}

func (s *Server) handleScroll(reader *lksdk.TextStreamReader, participant string) {
	delta, err := strconv.Atoi(reader.ReadAll())
	if err != nil || participant != s.state.current {
		return
	}

	direction := "clockwise"
	if delta < 0 {
		direction = "counterclockwise"
	}

	s.logger.Info("received scroll message", "delta", delta, "dir", direction, "user", participant)
}
