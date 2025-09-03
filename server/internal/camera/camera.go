package camera

import (
	"fmt"

	"github.com/pion/mediadevices"
	"github.com/pion/mediadevices/pkg/codec/x264"
	_ "github.com/pion/mediadevices/pkg/driver/camera"
	"github.com/pion/mediadevices/pkg/prop"
	"github.com/pion/webrtc/v4"
)

type Rig struct {
	Track  *mediadevices.VideoTrack
	Reader mediadevices.RTPReadCloser
}

func Setup() (*Rig, error) {
	params, err := x264.NewParams()
	if err != nil {
		return nil, fmt.Errorf("failed to generate new x264 codec params: %s", err)
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
		return nil, fmt.Errorf("unable to get user media: %s", err)
	}

	tracks := stream.GetVideoTracks()
	if len(tracks) < 1 {
		return nil, fmt.Errorf("no video tracks found")
	}

	track := tracks[0].(*mediadevices.VideoTrack)
	reader, err := track.NewRTPReader(webrtc.MimeTypeH264, 12345, 1200)
	if err != nil {
		return nil, fmt.Errorf("unable to convert track to reader: %s", err)
	}

	return &Rig{Track: track, Reader: reader}, nil
}
