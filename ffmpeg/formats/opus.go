package formats

import "strconv"

// OpusFormat contains additional ffmpeg fields specific to libopus
type OpusFormat struct {
	PacketLoss    int    // Expected packet loss percentage (-packet_loss)
	FrameDuration int    // Frame duration in ms (-frame_duration)
	VBR           string // Variable bit rate (-vbr)
	Application   string // Intended application type(-application)
}

// DiscordOpusFormat returns an OpusFormat with sane defaults for discord audio.
func DiscordOpusFormat() OpusFormat {
	return OpusFormat{
		PacketLoss:    1,
		FrameDuration: 20,
		VBR:           "on",
		Application:   "audio",
	}
}

func (o OpusFormat) Name() []string {
	return []string{"-c:a", "libopus"}
}

func (o OpusFormat) Format() string {
	return "ogg"
}

func (o OpusFormat) Args() []string {
	return []string{
		"-packet_loss", strconv.Itoa(o.PacketLoss),
		"-frame_duration", strconv.Itoa(o.FrameDuration),
		"-vbr", o.VBR,
		"-application", o.Application,
	}
}
