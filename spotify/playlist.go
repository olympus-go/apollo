package spotify

import (
	"github.com/eolso/librespot-golang/Spotify"
)

type Playlist struct {
	id              string
	spotifyPlaylist *Spotify.SelectedListContent
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

func (p Playlist) Tracks() []string {
	var tracks []string
	if p.spotifyPlaylist.Contents != nil {
		for _, item := range p.spotifyPlaylist.Contents.Items {
			trackUri := NewUri(item.GetUri())
			tracks = append(tracks, trackUri.Path)
		}
	}
	return tracks
}
