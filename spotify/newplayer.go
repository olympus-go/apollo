package spotify

import (
	"fmt"
	"github.com/eolso/librespot-golang/Spotify"
	"github.com/eolso/librespot-golang/librespot"
	"github.com/eolso/librespot-golang/librespot/core"
	"github.com/eolso/librespot-golang/librespot/utils"
	"github.com/rs/zerolog/log"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

var targetCodecs = []Spotify.AudioFile_Format{
	Spotify.AudioFile_OGG_VORBIS_320,
	Spotify.AudioFile_OGG_VORBIS_160,
	Spotify.AudioFile_OGG_VORBIS_96,
}

type Playerr struct {
	TrackQueue []Track

	config    PlayerConfig
	session   *core.Session
	eventChan chan PlayerEvent
	errChan   chan error

	trackQueue       chan Track
	downloadedTracks chan []byte
}

func NewPlayerr(config PlayerConfig) *Playerr {
	player := Playerr{
		config:           config,
		eventChan:        make(chan PlayerEvent),
		errChan:          make(chan error),
		downloadedTracks: make(chan []byte, 3),
	}

	if err := os.MkdirAll(config.ConfigHomeDir, 0755); err != nil {
		log.Warn().Err(err).Msg("failed to create config home directory")
	}
	if err := os.MkdirAll(config.CacheDir, 0755); err != nil {
		log.Warn().Err(err).Msg("failed to create song cache directory")
	}

	// TODO make this configurable
	go player.errorManager()

	return &player
}

func (p *Playerr) Login() error {
	if p.session != nil {
		return ErrPlayerAlreadyLoggedIn
	}

	// Check if auth token already exists
	authBytes, err := os.ReadFile(filepath.Join(p.config.ConfigHomeDir, "auth.token"))
	if err == nil && len(authBytes) != 0 {
		p.session, err = librespot.LoginSaved("asdf", authBytes, "georgetuney")
		return err
	}

	p.session, err = librespot.LoginOAuth("georgetuney", os.Getenv("SPOTIFY_ID"), os.Getenv("SPOTIFY_SECRET"))
	if err != nil {
		return fmt.Errorf("failed to initialize spotify client: %w", err)
	}

	err = ioutil.WriteFile(filepath.Join(p.config.ConfigHomeDir, "auth.token"), p.session.ReusableAuthBlob(), 0600)
	if err != nil {
		log.Warn().Err(err).Msg("failed to write auth token to filesystem")
	}

	return nil
}

func (p *Playerr) SearchTrack(query string, limit int) ([]Track, error) {
	var tracks []Track

	uri, ok := ConvertLinkToUri(query)

	if ok && uri.Authority == TrackResourceType {
		track, err := p.GetTrackById(uri.Path)
		if err != nil {
			return nil, err
		}
		tracks = append(tracks, track)
	} else {
		searchResponse, err := p.session.Mercury().Search(query, limit, p.session.Country(), p.session.Username())
		if err != nil {
			return nil, err
		}

		for _, metadataTrack := range searchResponse.Results.Tracks.Hits {
			trackUri := NewUri(metadataTrack.Uri)
			track, err := p.GetTrackById(trackUri.Path)
			if err != nil {
				return nil, err
			}
			tracks = append(tracks, track)
		}
	}

	return tracks, nil
}

func (p *Playerr) GetTrackById(id string) (Track, error) {
	track, err := p.session.Mercury().GetTrack(utils.Base62ToHex(id))
	return Track{spotifyTrack: track}, err
}

// Search is a helper function that will is return a list of tracks. If the query is a spotify URI, then it will return
// the relevant songs the link. If query is a simple string, it will return a track list from whatever the top hit was.
func (p *Playerr) Search(query string, limit int) ([]Track, error) {
	uri, _ := ConvertLinkToUri(query)
	switch uri.Authority {
	case ArtistResourceType:
		artists, err := p.SearchArtist(query, limit)
		if err != nil {
			return nil, err
		}
		if len(artists) == 0 {
			return nil, fmt.Errorf("query found no results")
		}

		var tracks []Track
		for _, trackId := range artists[0].TopTracks() {
			track, err := p.GetTrackById(trackId)
			if err != nil {
				return nil, err
			}
			tracks = append(tracks, track)
			if len(tracks) == limit {
				break
			}
		}
		return tracks, nil
	case PlaylistResourceType:
		playlists, err := p.SearchPlaylist(query, limit)
		if err != nil {
			return nil, err
		}
		if len(playlists) == 0 {
			return nil, fmt.Errorf("query found no results")
		}

		var tracks []Track
		for _, trackId := range playlists[0].Tracks() {
			track, err := p.GetTrackById(trackId)
			if err != nil {
				return nil, err
			}
			tracks = append(tracks, track)
			if len(tracks) == limit {
				break
			}
		}

		return tracks, nil
	// Just run a SearchTrack() regardless if it's a TrackResourceType or anything else
	default:
		tracks, err := p.SearchTrack(query, limit)
		if err != nil {
			return nil, err
		}

		if len(tracks) > limit {
			return tracks[:limit], nil
		} else {
			return tracks, nil
		}
	}
}

func (p *Playerr) SearchArtist(query string, limit int) ([]Artist, error) {
	var artists []Artist

	uri, ok := ConvertLinkToUri(query)

	if ok && uri.Authority == ArtistResourceType {
		artist, err := p.GetArtistById(uri.Path)
		if err != nil {
			return nil, err
		}
		artists = append(artists, artist)
	} else {
		searchResponse, err := p.session.Mercury().Search(query, limit, p.session.Country(), p.session.Username())
		if err != nil {
			return nil, err
		}

		for _, metadataArtist := range searchResponse.Results.Artists.Hits {
			artistUri := NewUri(metadataArtist.Uri)
			artist, err := p.GetArtistById(artistUri.Path)
			if err != nil {
				return nil, err
			}
			artists = append(artists, artist)
		}
	}

	return artists, nil
}

func (p *Playerr) GetArtistById(id string) (Artist, error) {
	artist, err := p.session.Mercury().GetArtist(utils.Base62ToHex(id))
	return Artist{spotifyArtist: artist}, err
}

func (p *Playerr) SearchPlaylist(query string, limit int) ([]Playlist, error) {
	var playlists []Playlist

	uri, ok := ConvertLinkToUri(query)

	if ok && uri.Authority == PlaylistResourceType {
		playlist, err := p.GetPlaylistById(uri.Path)
		if err != nil {
			return nil, err
		}
		playlists = append(playlists, playlist)
	} else {
		searchResponse, err := p.session.Mercury().Search(query, limit, p.session.Country(), p.session.Username())
		if err != nil {
			return nil, err
		}

		for _, metadataPlaylist := range searchResponse.Results.Playlists.Hits {
			playlistUri := NewUri(metadataPlaylist.Uri)
			playlist, err := p.GetPlaylistById(playlistUri.Path)
			if err != nil {
				return nil, err
			}
			playlists = append(playlists, playlist)
		}
	}

	return playlists, nil
}

func (p *Playerr) GetPlaylistById(id string) (Playlist, error) {
	playlist, err := p.session.Mercury().GetPlaylist(id)
	if err != nil {
		return Playlist{}, err
	}

	return Playlist{id: id, spotifyPlaylist: playlist}, nil
}

func (p *Playerr) QueueTrack(track Track) {
	p.TrackQueue = append(p.TrackQueue, track)
}

func (p *Playerr) DownloadTrack(track Track) (io.Reader, error) {
	var selectedFile *Spotify.AudioFile
	for _, file := range track.spotifyTrack.GetFile() {
		for _, codec := range targetCodecs {
			if file.GetFormat() == codec {
				selectedFile = file
				break
			}
		}
		if selectedFile != nil {
			break
		}
	}
	if selectedFile == nil {
		return nil, fmt.Errorf("failed to fetch track data %s", track.Id())
	}

	return p.session.Player().LoadTrack(selectedFile, track.spotifyTrack.GetGid())
}

func (p *Playerr) Play() {

}

func (p *Playerr) Pause() {

}

func (p *Playerr) Stop() {

}

func (p *Playerr) Next() {

}

func (p *Playerr) Previous() {

}

func (p *Playerr) Status() {

}

func (p *Playerr) OutStream() <-chan []byte {
	return p.downloadedTracks
}

/*
	QueueTrack(t Track) - appends to a queue of tracks. this should be a slice to make it modifiable?

	a worker goroutine should be running that's doing all the work under the hood. Listens on a command channel for updates
	like play, pause, next etc. it should also be in charge of the sending the loaded data into the outstream

*/
//
//func (p *Playerr) worker() {
//
//	for {
//		select {
//		case event := <-p.eventChan:
//
//		}
//	}
//}

//func (p *Playerr) downloadManager() {
//
//}

//func (p *Playerr) queueManager() {
//	for {
//		if len(p.TrackQueue) > 0 {
//			data, err := p.DownloadTrack(p.TrackQueue[0])
//			// Shrink before handling the error so things don't explode
//			p.TrackQueue = p.TrackQueue[1:]
//			if err != nil {
//				p.errChan <- err
//			}
//			p.downloadedTracks <- data
//		}
//	}
//}

func (p *Playerr) errorManager() {
	for err := range p.errChan {
		if err != nil {
			log.Error().Err(err).Msg("")
		}
	}
}
