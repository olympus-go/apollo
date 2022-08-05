package spotify

import (
	"bytes"
	"fmt"
	"github.com/rs/zerolog/log"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"testing"
)

var player *Playerr
var r io.Reader

func TestPlayerr_Login(t *testing.T) {
	player = NewPlayerr(DefaultPlayerConfig())
	if err := player.Login(); err != nil {
		t.Fatal(err)
	}

	log.Info().Msg("hi")
	tracks, _ := player.Search("https://open.spotify.com/playlist/6aShjzELZBXUtHwMDfnakd?si=428e238063ce4fad", 10)
	for _, track := range tracks {
		fmt.Println(track.Name())
	}
	//r, _ := player.DownloadTrack(tracks[0])
	//b, _ := ioutil.ReadAll(r)

	//os.WriteFile("livingroomsong.vorbis", b, 0755)
	//b, err := os.ReadFile("livingroomsong.vorbis")
	//if err != nil {
	//	t.Fatal(err)
	//}
	//
	//if _, err := encode(b); err != nil {
	//	t.Fatal(err)
	//}
}

//func TestPlayerr_DownloadTrack(t *testing.T) {
//	artist, _ := player.SearchArtist("https://open.spotify.com/artist/1L3hqVCHSL1Ajy3m0z1bAT?si=235f4c29b58b42e4", 1)
//	for _, trackId := range artist[0].TopTracks() {
//		track, _ := player.GetTrackById(trackId)
//		fmt.Println(track.Name())
//		//r, _ = player.DownloadTrack(track)
//		//break
//	}
//	//time.Sleep(6 * time.Second)
//}

//func TestPLayerr_HeheTrack(t *testing.T) {
//	sr, _ := player.session.Mercury().Search("mark hoppus magnolia park", 10, player.session.Country(), player.session.Username())
//
//	fmt.Println(sr.Results.Tracks)
//
//}

func encode(songData []byte) ([][]byte, error) {
	const (
		channels  int = 2                   // 1 for mono, 2 for stereo
		frameRate int = 48000               // audio sampling rate
		frameSize int = 960                 // uint16 size of each audio frame
		maxBytes  int = (frameSize * 2) * 2 // max size of opus data
	)

	//cmd := exec.Command("ffmpeg", "-i", "-", "-c:a", "libopus", "livingroomsong.opus")
	cmd := exec.Command("ffmpeg", "-i", "-", "-c:a", "libopus", "-f", "opus", "pipe:1")
	cmd.Stdin = bytes.NewReader(songData)

	ffmpegout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	//
	//ffmpegbuf := bufio.NewReader(ffmpegout)

	err = cmd.Start()
	if err != nil {
		return nil, err
	}
	//err = cmd.Wait()
	//if err != nil {
	//	return nil, err
	//}

	b, _ := ioutil.ReadAll(ffmpegout)

	os.WriteFile("./livingroomsong.opus", b, 0755)

	//var encodedBytes [][]byte
	//for {
	//	audiobuf := make([]int16, frameSize*channels)
	//	err = binary.Read(ffmpegbuf, binary.LittleEndian, &audiobuf)
	//	if err == io.EOF || err == io.ErrUnexpectedEOF {
	//		break
	//	}
	//	if err != nil {
	//		return nil, err
	//	}
	//
	//		opus, err := opusEncoder.Encode(audiobuf, frameSize, maxBytes)
	//		if err != nil {
	//			return nil, err
	//		}
	//
	//		encodedBytes = append(encodedBytes, opus)
	//}
	return nil, nil
}
