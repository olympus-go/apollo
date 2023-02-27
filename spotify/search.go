package spotify

import (
	"github.com/eolso/librespot-golang/librespot/metadata"
)

type Search struct {
	session *Session
	results *metadata.SearchResponse
	query   string
	uri     Uri
	isUri   bool
	shuffle bool
	limit   int
}

func (s *Search) Query(query string) *Search {
	s.query = query
	return s
}

func (s *Search) Limit(limit int) *Search {
	s.limit = limit
	return s
}

func (s *Search) Shuffle() *Search {
	s.shuffle = true
	return s
}

func (s *Search) TrackIds() ([]string, error) {
	if err := s.run(); err != nil {
		return nil, err
	}

	if s.isUri {
		return []string{s.uri.Path}, nil
	}

	trackIds := make([]string, 0, s.limit)
	for i, metadataTrack := range s.results.Results.Tracks.Hits {
		if i == s.limit {
			break
		}

		trackUri := NewUri(metadataTrack.Uri)
		trackIds = append(trackIds, trackUri.Path)
	}

	return trackIds, nil
}

func (s *Search) Tracks() ([]Track, error) {
	trackIds, err := s.TrackIds()
	if err != nil {
		return nil, err
	}

	tracks := make([]Track, 0, len(trackIds))
	for _, trackId := range trackIds {
		track, err := s.session.GetTrackById(trackId)
		if err != nil {
			return nil, err
		}

		tracks = append(tracks, track)
	}

	return tracks, nil
}

func (s *Search) ArtistIds() ([]string, error) {
	if err := s.run(); err != nil {
		return nil, err
	}

	if s.isUri {
		return []string{s.uri.Path}, nil
	}

	artistIds := make([]string, 0, s.limit)
	for i, metadataTrack := range s.results.Results.Artists.Hits {
		if i == s.limit {
			break
		}

		artistUri := NewUri(metadataTrack.Uri)
		artistIds = append(artistIds, artistUri.Path)
	}

	return artistIds, nil
}

func (s *Search) Artists() ([]Artist, error) {
	artistIds, err := s.ArtistIds()
	if err != nil {
		return nil, err
	}

	artists := make([]Artist, 0, len(artistIds))
	for _, artistId := range artistIds {
		artist, err := s.session.GetArtistById(artistId)
		if err != nil {
			return nil, err
		}

		artists = append(artists, artist)
	}

	return artists, nil
}

func (s *Search) PlaylistIds() ([]string, error) {
	if err := s.run(); err != nil {
		return nil, err
	}

	if s.isUri {
		return []string{s.uri.Path}, nil
	}

	playlistIds := make([]string, 0, s.limit)
	for i, metadataTrack := range s.results.Results.Playlists.Hits {
		if i == s.limit {
			break
		}

		playlistUri := NewUri(metadataTrack.Uri)
		playlistIds = append(playlistIds, playlistUri.Path)
	}

	return playlistIds, nil
}

func (s *Search) Playlists() ([]Playlist, error) {
	playlistIds, err := s.PlaylistIds()
	if err != nil {
		return nil, err
	}

	playlists := make([]Playlist, 0, len(playlistIds))
	for _, playlistId := range playlistIds {
		playlist, err := s.session.GetPlaylistById(playlistId)
		if err != nil {
			return nil, err
		}

		playlists = append(playlists, playlist)
	}

	return playlists, nil
}

func (s *Search) run() error {
	s.uri, s.isUri = ConvertLinkToUri(s.query)

	var err error
	s.results, err = s.session.client.Mercury().Search(s.query,
		s.limit,
		s.session.client.Country(),
		s.session.client.Username(),
	)

	return err
}
