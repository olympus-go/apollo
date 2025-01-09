package spotify

import (
	"encoding/hex"
	"fmt"
	"io"
	"time"

	"github.com/eolso/librespot-golang/Spotify"
	"github.com/eolso/librespot-golang/librespot/player"
	"github.com/eolso/librespot-golang/librespot/utils"
)

type Track struct {
	spotifyTrack *Spotify.Track
	player       *player.Player

	customName        string
	customArtist      string
	customDescription string
	customAlbum       string
	customImage       string
}

func (t *Track) Name() string {
	if len(t.customName) > 0 {
		return t.customName
	}

	return t.spotifyTrack.GetName()
}

func (t *Track) Artist() string {
	if len(t.customArtist) > 0 {
		return t.customArtist
	}

	if len(t.spotifyTrack.Artist) == 0 {
		return "Unknown"
	}

	return t.spotifyTrack.Artist[0].GetName()
}

func (t *Track) Metadata() map[string]string {
	return nil
}

func (t *Track) Id() string {
	return utils.ConvertTo62(t.spotifyTrack.GetGid())
}

func (t *Track) Description() string {
	if len(t.customDescription) > 0 {
		return t.customDescription
	}

	return fmt.Sprintf("%s by %s", t.Name(), t.Artist())
}

func (t *Track) Album() string {
	if len(t.customAlbum) > 0 {
		return t.customAlbum
	}

	if t.spotifyTrack.Album == nil {
		return "Unknown"
	}

	return t.spotifyTrack.GetAlbum().GetName()
}

func (t *Track) Image() string {
	if len(t.customImage) > 0 {
		return t.customImage
	}

	image := t.spotifyTrack.GetAlbum().GetCoverGroup().GetImage()
	if len(image) > 0 {
		return fmt.Sprintf("https://i.scdn.co/image/%032s", hex.EncodeToString(image[0].GetFileId()))
	}
	return ""
}

func (t *Track) Duration() time.Duration {
	return time.Duration(t.spotifyTrack.GetDuration()) * time.Millisecond
}

func (t *Track) Type() string {
	return "spotify track"
}

func (t *Track) Download() (io.ReadCloser, error) {
	var selectedFile *Spotify.AudioFile

	audioFiles := t.spotifyTrack.GetFile()

	// If the track returned no audio files, try the alternatives until files are found
	if len(audioFiles) == 0 {
		for _, alternative := range t.spotifyTrack.Alternative {
			audioFiles = alternative.GetFile()
			if len(audioFiles) > 0 {
				break
			}
		}
	}

	// All alternatives tried, still no files
	if len(audioFiles) == 0 {
		return nil, fmt.Errorf("failed to fetch track data %s", t.Id())
	}

	// Try and grab a desired codec first
	for _, file := range t.spotifyTrack.GetFile() {
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

	// Grab whatever is left
	if selectedFile == nil {
		selectedFile = audioFiles[0]
	}

	return t.player.LoadTrack(selectedFile, t.spotifyTrack.GetGid())
}

func (t *Track) SetCustomName(name string) {
	t.customName = name
}

func (t *Track) SetCustomArtist(artist string) {
	t.customArtist = artist
}

func (t *Track) SetCustomDescription(description string) {
	t.customDescription = description
}

func (t *Track) SetCustomAlbum(album string) {
	t.customAlbum = album
}

func (t *Track) SetCustomImage(image string) {
	t.customImage = image
}
