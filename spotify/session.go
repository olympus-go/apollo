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

func (s *Session) GetTrackById(id string) (Track, error) {
	track, err := s.client.Mercury().GetTrack(utils.Base62ToHex(id))
	return Track{spotifyTrack: track, player: s.client.Player()}, err
}

func (s *Session) GetArtistById(id string) (Artist, error) {
	artist, err := s.client.Mercury().GetArtist(utils.Base62ToHex(id))
	return Artist{spotifyArtist: artist, session: s}, err
}

func (s *Session) GetPlaylistById(id string) (Playlist, error) {
	playlist, err := s.client.Mercury().GetPlaylist(id)
	if err != nil {
		return Playlist{}, err
	}

	return Playlist{id: id, spotifyPlaylist: playlist, session: s}, nil
}

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
