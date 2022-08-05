package spotify

import (
	"encoding/hex"
	"fmt"
	"github.com/eolso/librespot-golang/Spotify"
	"github.com/eolso/librespot-golang/librespot/utils"
)

type Track struct {
	spotifyTrack *Spotify.Track
}

func (t *Track) Id() string {
	return utils.ConvertTo62(t.spotifyTrack.GetGid())
}

func (t *Track) Name() string {
	return t.spotifyTrack.GetName()
}

func (t *Track) Artist() string {
	if len(t.spotifyTrack.Artist) == 0 {
		return "Unknown"
	}

	return t.spotifyTrack.Artist[0].GetName()
}

func (t *Track) Image() string {
	image := t.spotifyTrack.GetAlbum().GetCoverGroup().GetImage()
	if len(image) > 0 {
		return fmt.Sprintf("https://i.scdn.co/image/%032s", hex.EncodeToString(image[0].GetFileId()))
	}
	return ""
}

func (t *Track) Duration() int32 {
	return t.spotifyTrack.GetDuration()
}
