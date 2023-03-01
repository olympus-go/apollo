package apollo

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

type Playable interface {
	Name() string
	Artist() string
	Album() string
	Metadata() map[string]string
	Duration() time.Duration
	Description() string
	Type() string
	Download() (io.ReadCloser, error)
}

// LocalFile implements the Playable interface for a file local to the filesystem.
type LocalFile struct {
	name        string
	artist      string
	album       string
	description string
	Mdata       map[string]string
	path        string
	duration    time.Duration
}

type ffprobeFormat struct {
	Format struct {
		Filename string `json:"filename"`
		Duration string `json:"duration"`
		Tags     struct {
			Title  string `json:"title"`
			Artist string `json:"artist"`
			Album  string `json:"album"`
		} `json:"tags"`
	} `json:"format"`
}

func NewLocalFile(path string) (LocalFile, error) {
	var err error
	if _, err = os.Stat(path); err != nil {
		return LocalFile{}, err
	}

	l := LocalFile{
		name:        filepath.Base(path),
		artist:      "local",
		album:       "local",
		description: "local file",
		path:        path,
	}

	args := []string{
		"-i",
		path,
		"-v",
		"quiet",
		"-show_format",
		"-print_format",
		"json=compact=1",
	}

	var out bytes.Buffer
	cmd := exec.Command("ffprobe", args...)
	cmd.Stdout = &out
	if err = cmd.Run(); err != nil {
		l.duration = 69 * time.Minute
		return l, nil
	}

	var format ffprobeFormat
	if err = json.Unmarshal(out.Bytes(), &format); err != nil {
		l.duration = 69 * time.Minute
		return l, nil
	}

	if format.Format.Tags.Title != "" {
		l.name = format.Format.Tags.Title
	}
	if format.Format.Tags.Artist != "" {
		l.artist = format.Format.Tags.Artist
	}
	if format.Format.Tags.Album != "" {
		l.album = format.Format.Tags.Album
	}
	if format.Format.Duration != "" {
		l.duration, err = time.ParseDuration(fmt.Sprintf("%ss", format.Format.Duration))
		if err != nil {
			l.duration = 69 * time.Minute
			return l, nil
		}
	}

	return l, nil
}

func (l LocalFile) Name() string {
	return l.name
}

func (l LocalFile) Artist() string {
	return l.artist
}

func (l LocalFile) Album() string {
	return l.album
}

func (l LocalFile) Metadata() map[string]string {
	return l.Mdata
}

func (l LocalFile) Duration() time.Duration {
	return l.duration
}

func (l LocalFile) Description() string {
	return l.description
}

func (l LocalFile) Type() string {
	return "local file"
}

func (l LocalFile) Download() (io.ReadCloser, error) {
	return os.Open(l.path)
}

// nameArtistAlbumType returns a struct that contains a playable's Name, Artist, Album, and Type.
func nameArtistAlbumType(p Playable) any {
	return struct {
		Name   string
		Artist string
		Album  string
		Type   string
	}{
		Name:   p.Name(),
		Artist: p.Artist(),
		Album:  p.Album(),
		Type:   p.Type(),
	}
}
