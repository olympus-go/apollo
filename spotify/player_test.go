package spotify

import (
	"fmt"
	"github.com/rs/zerolog/log"
	"io"
	"testing"
)

var player *Player
var r io.Reader

func TestPlayerr_Login(t *testing.T) {
	player = NewPlayer(DefaultPlayerConfig())
	if err := player.Login(); err != nil {
		t.Fatal(err)
	}

	log.Info().Msg("hi")
	tracks, _ := player.Search("https://open.spotify.com/playlist/6aShjzELZBXUtHwMDfnakd?si=428e238063ce4fad", 10)
	for _, track := range tracks {
		fmt.Println(track.Name())
	}
}
