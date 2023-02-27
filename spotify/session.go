package spotify

import (
	"os"
	"path/filepath"

	"github.com/eolso/librespot-golang/Spotify"
	"github.com/eolso/librespot-golang/librespot"
	"github.com/eolso/librespot-golang/librespot/core"
	"github.com/eolso/librespot-golang/librespot/utils"
	"github.com/rs/zerolog/log"
)

// targetCodecs sets the order priority of codecs to fetch. TODO enable setting this.
var targetCodecs = []Spotify.AudioFile_Format{
	Spotify.AudioFile_OGG_VORBIS_320,
	Spotify.AudioFile_OGG_VORBIS_160,
	Spotify.AudioFile_OGG_VORBIS_96,
	Spotify.AudioFile_MP3_320,
	Spotify.AudioFile_MP3_256,
	Spotify.AudioFile_MP3_160,
	Spotify.AudioFile_MP3_160_ENC,
	Spotify.AudioFile_MP3_96,
	Spotify.AudioFile_AAC_320,
	Spotify.AudioFile_AAC_160,
	Spotify.AudioFile_MP4_128,
	Spotify.AudioFile_MP4_128_DUAL,
	Spotify.AudioFile_OTHER5,
	Spotify.AudioFile_OTHER3,
}

// Session is the base object used for interacting with spotify. All auth and api calls go through Session one way or
// another.
type Session struct {
	config SessionConfig
	client *core.Session
}

func NewSession(config SessionConfig) *Session {
	session := Session{
		config: config,
	}

	if config.ConfigHomeDir != "" {
		if err := os.MkdirAll(config.ConfigHomeDir, 0755); err != nil {
			log.Warn().Err(err).Msg("failed to create config home directory")
		}
		if err := os.MkdirAll(config.CacheDir, 0755); err != nil {
			log.Warn().Err(err).Msg("failed to create song cache directory")
		}
	}

	return &session
}

func (s *Session) Login(deviceName string) error {
	if s.LoggedIn() {
		return ErrPlayerAlreadyLoggedIn
	}

	// Check if auth token already exists
	if s.config.ConfigHomeDir != "" {
		authBytes, err := os.ReadFile(filepath.Join(s.config.ConfigHomeDir, "auth.token"))
		if err == nil && len(authBytes) != 0 {
			// TODO make username overrideable
			s.client, err = librespot.LoginSaved("apollo", authBytes, deviceName)
			return err
		} else {
			return ErrTokenNotFound
		}
	} else {
		return ErrTokenNotFound
	}

	// TODO separate this out
	//var err error
	//s.client, err = librespot.LoginOAuth(deviceName, os.Getenv("SPOTIFY_ID"), os.Getenv("SPOTIFY_SECRET"), s.config.OAuthCallback)
	//if err != nil {
	//	return fmt.Errorf("failed to initialize spotify client: %w", err)
	//}
	//
	//if s.config.ConfigHomeDir != "" {
	//	err = os.WriteFile(filepath.Join(s.config.ConfigHomeDir, "auth.token"), s.client.ReusableAuthBlob(), 0600)
	//	if err != nil {
	//		log.Warn().Err(err).Msg("failed to write auth token to filesystem")
	//	}
	//}

	return nil
}

func (s *Session) LoginWithToken(deviceName string, token string) error {
	var err error
	s.client, err = core.LoginOAuthToken(token, deviceName)

	if s.config.ConfigHomeDir != "" {
		err = os.WriteFile(filepath.Join(s.config.ConfigHomeDir, "auth.token"), s.client.ReusableAuthBlob(), 0600)
		if err != nil {
			log.Warn().Err(err).Msg("failed to write auth token to filesystem")
		}
	}

	return err
}

//func (s *Session) SearchTrack(query string, limit int) ([]Track, error) {
//	var tracks []Track
//
//	uri, ok := ConvertLinkToUri(query)
//
//	if ok && uri.Authority == TrackResourceType {
//		track, err := s.GetTrackById(uri.Path)
//		if err != nil {
//			return nil, err
//		}
//		tracks = append(tracks, track)
//	} else {
//		searchResponse, err := s.client.Mercury().Search(query, limit, s.client.Country(), s.client.Username())
//		if err != nil {
//			return nil, err
//		}
//
//		for _, metadataTrack := range searchResponse.Results.Tracks.Hits {
//			trackUri := NewUri(metadataTrack.Uri)
//			track, err := s.GetTrackById(trackUri.Path)
//			if err != nil {
//				return nil, err
//			}
//			tracks = append(tracks, track)
//		}
//	}
//
//	return tracks, nil
//}

func (s *Session) GetTrackById(id string) (Track, error) {
	track, err := s.client.Mercury().GetTrack(utils.Base62ToHex(id))
	return Track{spotifyTrack: track, player: s.client.Player()}, err
}

// Search is a helper function that will is return a list of tracks. If the query is a spotify URI, then it will return
// the relevant songs the link. If query is a simple string, it will return a track list from whatever the top hit was.
//func (s *Session) Search(query string, limit int) ([]Track, error) {
//	uri, _ := ConvertLinkToUri(query)
//	switch uri.Authority {
//	case ArtistResourceType:
//		artists, err := s.SearchArtist(query, limit)
//		if err != nil {
//			return nil, err
//		}
//		if len(artists) == 0 {
//			return nil, fmt.Errorf("query found no results")
//		}
//
//		var tracks []Track
//		for _, trackId := range artists[0].TopTrackIds() {
//			track, err := s.GetTrackById(trackId)
//			if err != nil {
//				return nil, err
//			}
//			tracks = append(tracks, track)
//			if len(tracks) == limit {
//				break
//			}
//		}
//		return tracks, nil
//	case PlaylistResourceType:
//		playlists, err := s.SearchPlaylist(query, limit)
//		if err != nil {
//			return nil, err
//		}
//		if len(playlists) == 0 {
//			return nil, fmt.Errorf("query found no results")
//		}
//
//		var tracks []Track
//		for _, trackId := range playlists[0].TrackIds() {
//			track, err := s.GetTrackById(trackId)
//			if err != nil {
//				return nil, err
//			}
//			tracks = append(tracks, track)
//			if len(tracks) == limit {
//				break
//			}
//		}
//
//		return tracks, nil
//	// Just run a SearchTrack() regardless if it's a TrackResourceType or anything else
//	default:
//		tracks, err := s.SearchTrack(query, limit)
//		if err != nil {
//			return nil, err
//		}
//
//		if len(tracks) > limit {
//			return tracks[:limit], nil
//		} else {
//			return tracks, nil
//		}
//	}
//}

//func (s *Session) SearchArtist(query string, limit int) ([]Artist, error) {
//	var artists []Artist
//
//	uri, ok := ConvertLinkToUri(query)
//
//	if ok && uri.Authority == ArtistResourceType {
//		artist, err := s.GetArtistById(uri.Path)
//		if err != nil {
//			return nil, err
//		}
//		artists = append(artists, artist)
//	} else {
//		searchResponse, err := s.client.Mercury().Search(query, limit, s.client.Country(), s.client.Username())
//		if err != nil {
//			return nil, err
//		}
//
//		for _, metadataArtist := range searchResponse.Results.Artists.Hits {
//			artistUri := NewUri(metadataArtist.Uri)
//			artist, err := s.GetArtistById(artistUri.Path)
//			if err != nil {
//				return nil, err
//			}
//			artists = append(artists, artist)
//		}
//	}
//
//	return artists, nil
//}

func (s *Session) GetArtistById(id string) (Artist, error) {
	artist, err := s.client.Mercury().GetArtist(utils.Base62ToHex(id))
	return Artist{spotifyArtist: artist, session: s}, err
}

//func (s *Session) SearchPlaylist(query string, limit int) ([]Playlist, error) {
//	var playlists []Playlist
//
//	uri, ok := ConvertLinkToUri(query)
//
//	if ok && uri.Authority == PlaylistResourceType {
//		playlist, err := s.GetPlaylistById(uri.Path)
//		if err != nil {
//			return nil, err
//		}
//		playlists = append(playlists, playlist)
//	} else {
//		searchResponse, err := s.client.Mercury().Search(query, limit, s.client.Country(), s.client.Username())
//		if err != nil {
//			return nil, err
//		}
//
//		for _, metadataPlaylist := range searchResponse.Results.Playlists.Hits {
//			playlistUri := NewUri(metadataPlaylist.Uri)
//			playlist, err := s.GetPlaylistById(playlistUri.Path)
//			if err != nil {
//				return nil, err
//			}
//			playlists = append(playlists, playlist)
//		}
//	}
//
//	return playlists, nil
//}

func (s *Session) GetPlaylistById(id string) (Playlist, error) {
	playlist, err := s.client.Mercury().GetPlaylist(id)
	if err != nil {
		return Playlist{}, err
	}

	return Playlist{id: id, spotifyPlaylist: playlist, session: s}, nil
}

//func (s *Session) DownloadTrack(track Track) (io.Reader, error) {
//	var selectedFile *Spotify.AudioFile
//
//	audioFiles := track.spotifyTrack.GetFile()
//
//	// If the track returned no audio files, try the alternatives until files are found
//	if len(audioFiles) == 0 {
//		for _, alternative := range track.spotifyTrack.Alternative {
//			audioFiles = alternative.GetFile()
//			if len(audioFiles) > 0 {
//				break
//			}
//		}
//	}
//
//	// All alternatives tried, still no files
//	if len(audioFiles) == 0 {
//		return nil, fmt.Errorf("failed to fetch track data %s", track.Id())
//	}
//
//	// Try and grab a desired codec first
//	for _, file := range track.spotifyTrack.GetFile() {
//		for _, codec := range targetCodecs {
//			if file.GetFormat() == codec {
//				selectedFile = file
//				break
//			}
//		}
//		if selectedFile != nil {
//			break
//		}
//	}
//
//	// Grab whatever is left
//	if selectedFile == nil {
//		selectedFile = audioFiles[0]
//	}
//
//	return s.client.Player().LoadTrack(selectedFile, track.spotifyTrack.GetGid())
//}

func (s *Session) Username() string {
	if s.client != nil {
		return s.client.Username()
	}

	return ""
}

func (s *Session) LoggedIn() bool {
	return s.client != nil
}

func (s *Session) Search(query string) *Search {
	return &Search{
		session: s,
		query:   query,
		limit:   10,
	}
}
