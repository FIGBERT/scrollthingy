package server

import (
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"

	"github.com/livekit/protocol/auth"
)

func token(s *Server) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		at := auth.NewAccessToken(os.Getenv("LIVEKIT_API_KEY"), os.Getenv("LIVEKIT_API_SECRET"))
		grant := &auth.VideoGrant{
			RoomJoin: true,
			Room:     "default",
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
		w.Write([]byte(token))
	}
}
