package spotify

import (
	"encoding/hex"
	"fmt"

	"github.com/eolso/librespot-golang/Spotify"
	"github.com/eolso/librespot-golang/librespot/utils"
)

type Artist struct {
	spotifyArtist *Spotify.Artist
	session       *Session
}

func (a Artist) Id() string {
	return utils.ConvertTo62(a.spotifyArtist.GetGid())
}

func (a Artist) Name() string {
	return a.spotifyArtist.GetName()
}

func (a Artist) Bio() string {
	bio := a.spotifyArtist.GetBiography()
	if len(bio) > 0 {
		return bio[0].GetText()
	}
	return ""
}

func (a Artist) Image() string {
	image := a.spotifyArtist.GetPortraitGroup().GetImage()
	if len(image) > 0 {
		return fmt.Sprintf("https://i.scdn.co/image/%032s", hex.EncodeToString(image[0].GetFileId()))
	}
	return ""
}

func (a Artist) TopTrackIds() []string {
	topTracks := a.spotifyArtist.GetTopTrack()
	if len(topTracks) == 0 {
		return nil
	}

	var ids []string
	for _, track := range topTracks[0].GetTrack() {
		ids = append(ids, fmt.Sprintf("%s", utils.ConvertTo62(track.GetGid())))
	}

	return ids
}

func (a Artist) TopTracks() ([]Track, error) {
	trackIds := a.TopTrackIds()

	tracks := make([]Track, 0, len(trackIds))
	for _, trackId := range trackIds {
		track, err := a.session.GetTrackById(trackId)
		if err != nil {
			return nil, err
		}

		tracks = append(tracks, track)
	}

	return tracks, nil
}
