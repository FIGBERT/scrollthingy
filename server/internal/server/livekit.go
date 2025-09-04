package server

import (
	"context"
	"net/http"
	"os"
	"slices"

	"github.com/google/uuid"
	"github.com/pion/webrtc/v4"

	"github.com/livekit/protocol/auth"
	lksdk "github.com/livekit/server-sdk-go/v2"
)

const ROOM_NAME = "default"

func (s *Server) token() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		at := auth.NewAccessToken(os.Getenv("LIVEKIT_API_KEY"), os.Getenv("LIVEKIT_API_SECRET"))
		grant := &auth.VideoGrant{RoomJoin: true, Room: ROOM_NAME}
		at.SetVideoGrant(grant).SetIdentity(uuid.NewString())

		token, err := at.ToJWT()
		if err != nil {
			s.logger.Error("unable to convert access token to jwt", "err", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		url := os.Getenv("LIVEKIT_URL")
		if url == "" {
			s.logger.Error("LIVEKIT_URL is blank or undefined", "err", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:1234")
		w.Write([]byte(url + "\n" + token))
	}
}

func (s *Server) join_room() error {
	room, err := lksdk.ConnectToRoom(
		os.Getenv("LIVEKIT_URL"),
		lksdk.ConnectInfo{
			APIKey:              os.Getenv("LIVEKIT_API_KEY"),
			APISecret:           os.Getenv("LIVEKIT_API_SECRET"),
			RoomName:            ROOM_NAME,
			ParticipantIdentity: "server",
		},
		&lksdk.RoomCallback{
			OnParticipantConnected:    s.participantJoined,
			OnParticipantDisconnected: s.participantLeft,
		},
	)
	if err != nil {
		return err
	}

	s.room = room
	return nil
}

func (s *Server) participantJoined(user *lksdk.RemoteParticipant) {
	id := user.Identity()
	s.state.users = append(s.state.users, id)
	if s.state.current == "" {
		s.state.current = id
	}
}

func (s *Server) participantLeft(user *lksdk.RemoteParticipant) {
	id := user.Identity()
	if s.state.current == id {
		s.state.users = s.state.users[1:]
		if len(s.state.users) > 0 {
			s.state.current = s.state.users[0]
		} else {
			s.state.current = ""
		}
	} else {
		s.state.users = slices.Collect(func(yield func(string) bool) {
			for _, u := range s.state.users {
				if u != id {
					if !yield(u) {
						return
					}
				}
			}
		})
	}
}

func (s *Server) publish_camera() {
	if s.room == nil {
		s.logger.Error("no room available to publish feed. did you call join_room()?", "server", s)
		return
	}

	codec := webrtc.RTPCodecCapability{
		MimeType:  webrtc.MimeTypeH264,
		ClockRate: 90000,
	}

	track, err := lksdk.NewLocalTrack(codec)
	if err != nil {
		s.logger.Error("failed to create livekit track", "err", err)
		return
	}

	_, err = s.room.LocalParticipant.PublishTrack(track, &lksdk.TrackPublicationOptions{})
	if err != nil {
		s.logger.Error("publishing camera feed did not work as expected", "err", err)
		return
	}

	go s.writeRTPPackets(track)
}

func (s *Server) writeRTPPackets(track *lksdk.LocalTrack) {
	ctx := context.Background()
	for {
		select {
		case <-ctx.Done():
			return
		default:
			packets, release, err := s.rig.Reader.Read()
			if err != nil {
				s.logger.Error("failed to read RTP packets", "err", err)
				return
			}

			for _, packet := range packets {
				if err := track.WriteRTP(packet, nil); err != nil {
					s.logger.Error("failed to write RTP packet", "err", err)
					release()
					return
				}
			}
			release()
		}
	}
}
