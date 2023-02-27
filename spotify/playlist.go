package spotify

import (
	"encoding/hex"
	"fmt"

	"github.com/eolso/librespot-golang/Spotify"
)

type Playlist struct {
	id              string
	spotifyPlaylist *Spotify.SelectedListContent
	session         *Session
}

func (p Playlist) Id() string {
	return p.id
}

func (p Playlist) Name() string {
	return p.spotifyPlaylist.Attributes.GetName()
}

func (p Playlist) Description() string {
	return p.spotifyPlaylist.Attributes.GetDescription()
}

func (p Playlist) TrackIds() []string {
	var tracks []string
	if p.spotifyPlaylist.Contents != nil {
		for _, item := range p.spotifyPlaylist.Contents.Items {
			trackUri := NewUri(item.GetUri())
			tracks = append(tracks, trackUri.Path)
		}
	}
	return tracks
}

func (p Playlist) Tracks() ([]Track, error) {
	trackIds := p.TrackIds()
	if len(trackIds) == 0 {
		return nil, fmt.Errorf("no tracks")
	}

	tracks := make([]Track, 0, len(trackIds))
	for _, trackId := range trackIds {
		track, err := p.session.GetTrackById(trackId)
		if err != nil {
			return nil, err
		}
		tracks = append(tracks, track)
	}

	return tracks, nil
}

func (p Playlist) Image() string {
	image := p.spotifyPlaylist.GetAttributes().GetPicture()
	if len(image) > 0 {
		return fmt.Sprintf("https://i.scdn.co/image/%032s", hex.EncodeToString(image))
	}
	return ""
}
