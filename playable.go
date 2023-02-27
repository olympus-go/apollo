package apollo

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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
	name  string
	path  string
	Mdata map[string]string
}

func NewLocalFile(path string) (LocalFile, error) {
	if _, err := os.Stat(path); err != nil {
		return LocalFile{}, err
	}
	name := filepath.Base(path)

	return LocalFile{
		name: name,
		path: path,
	}, nil
}

func (l LocalFile) Name() string {
	return l.name
}

func (l LocalFile) Artist() string {
	return "local"
}

func (l LocalFile) Album() string {
	return "local"
}

func (l LocalFile) Metadata() map[string]string {
	return l.Mdata
}

func (l LocalFile) Duration() time.Duration {
	args := []string{
		"-i",
		l.path,
		"-show_entries",
		"format=duration",
		"-v",
		"quiet",
		"-of",
		"csv=p=0",
	}

	var out strings.Builder
	cmd := exec.Command("ffprobe", args...)
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return 69 * time.Minute
	}

	dur, err := time.ParseDuration(fmt.Sprintf("%ss", strings.TrimSpace(out.String())))
	if err != nil {
		return 69 * time.Minute
	}

	return dur
}

func (l LocalFile) Description() string {
	return "local file " + l.name
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
