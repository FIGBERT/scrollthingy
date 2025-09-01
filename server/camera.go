package main

import (
	"log/slog"

	"github.com/pion/mediadevices"
	"github.com/pion/mediadevices/pkg/codec/x264"
	_ "github.com/pion/mediadevices/pkg/driver/camera"
	"github.com/pion/mediadevices/pkg/prop"
	"github.com/pion/webrtc/v4"
)

func run(logger *slog.Logger) {
	params, err := x264.NewParams()
	if err != nil {
		logger.Error("failure to generate new codec params", "format", "x264", "err", err)
	}
	params.BitRate = 2_000_000

	stream, err := mediadevices.GetUserMedia(mediadevices.MediaStreamConstraints{
		Video: func(constraint *mediadevices.MediaTrackConstraints) {
			constraint.Width = prop.Int(600)
			constraint.Height = prop.Int(400)
		},
		Codec: mediadevices.NewCodecSelector(mediadevices.WithVideoEncoders(&params)),
	})
	if err != nil {
		logger.Error("unable to get user media", "err", err)
		return
	}

	tracks := stream.GetVideoTracks()
	if len(tracks) < 1 {
		logger.Error("no video tracks found")
		return
	}
	logger.Info("got video tracks", "tracks", tracks, "count", len(tracks))

	track := tracks[0].(*mediadevices.VideoTrack)
	defer track.Close()

	_, err = track.NewEncodedIOReader(webrtc.MimeTypeH264)
	if err != nil {
		logger.Error("unable to convert track to reader", "err", err, "track", track.ID())
	}
}
