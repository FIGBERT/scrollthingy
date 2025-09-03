package server

import (
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/pion/webrtc/v4"

	"github.com/livekit/protocol/auth"
	lksdk "github.com/livekit/server-sdk-go/v2"
)

const ROOM_NAME = "default"

func (s *Server) token() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		at := auth.NewAccessToken(os.Getenv("LIVEKIT_API_KEY"), os.Getenv("LIVEKIT_API_SECRET"))
		grant := &auth.VideoGrant{
			RoomJoin: true,
			Room:     ROOM_NAME,
		}
		at.
			SetVideoGrant(grant).
			SetValidFor(time.Hour).
			SetIdentity(uuid.NewString())

		token, err := at.ToJWT()
		if err != nil {
			s.logger.Error("unable to convert access token to jwt", "err", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:1234")
		w.Write([]byte(token))
	}
}

func join_room() (*lksdk.Room, error) {
	return lksdk.ConnectToRoom(
		os.Getenv("LIVEKIT_URL"),
		lksdk.ConnectInfo{
			APIKey:              os.Getenv("LIVEKIT_API_KEY"),
			APISecret:           os.Getenv("LIVEKIT_API_SECRET"),
			RoomName:            ROOM_NAME,
			ParticipantIdentity: "server",
		},
		&lksdk.RoomCallback{},
	)
}

func (s *Server) publish_camera() {
	if s.room == nil {
		s.logger.Error("no room available to publish feed. did you call join_room()?", "server", s)
		return
	}

	track, err := lksdk.NewLocalReaderTrack(s.rig.Reader, webrtc.MimeTypeH264)
	if err != nil {
		s.logger.Error("failed to create livekit track from camera rig reader", "err", err)
		return
	}

	_, err = s.room.LocalParticipant.PublishTrack(track, &lksdk.TrackPublicationOptions{})
	if err != nil {
		s.logger.Error("publishing camera feed did not work as expected", "err", err)
	}
}
