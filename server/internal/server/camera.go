package server

import "github.com/fcjr/scroll-together/server/internal/camera"

func (s *Server) setupCamera() error {
	rig, err := camera.Setup()
	if err != nil {
		return err
	}

	s.rig = rig
	return nil
}
