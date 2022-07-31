package spotify

// OLD FILE

import (
	"errors"
	"fmt"
	"github.com/eolso/librespot-golang/Spotify"
	"github.com/eolso/librespot-golang/librespot"
	"github.com/eolso/librespot-golang/librespot/core"
	"github.com/eolso/librespot-golang/librespot/metadata"
	"github.com/eolso/librespot-golang/librespot/utils"
	"github.com/rs/zerolog/log"
	"golang.org/x/net/context"
	"io/ioutil"
	"os"
	"strings"
	"time"
)

const (
	channels  = 2                   // 1 for mono, 2 for stereo
	frameRate = 48000               // audio sampling rate
	frameSize = 960                 // uint16 size of each audio frame
	maxBytes  = (frameSize * 2) * 2 // max size of opus data
	//defaultBufferSize = 25000000            // 25MB
)

//func init() {
//	if err := os.MkdirAll(songDataDir, 0755); err != nil {
//		log.Fatal().Err(err).Msg("failed to create temp dir")
//	}
//}

//{
//	{
//		[]
//		The Off-Season
//		spotify:album:4JAvwK4APPArjIsOdGoJXX
//	}
//	[
//		{
//			J. Cole
//			spotify:artist:6l3HvQ5sa6mXTsMTB19rO5
//		}
//	]
//	https://i.scdn.co/image/ab67616d00001e0210e6745bb2f179dd3616b85f a m a r i spotify:track:2cnKST6T9qUo2i907lm8zX 148421 72
//}

type State int8

const (
	StoppedState State = iota
	PausedState
	PlayingState
)

type TrackData struct {
	metadata.Track
	TimeElapsed float64
	//Sample      []byte
}

type trackInfo struct {
	metadata.Track
	data [][]byte
}

type Player struct {
	session          *core.Session
	state            State
	currentlyPlaying TrackData
	queueInfo        []metadata.Track
	isStarted        bool

	// trackQueue is a queue of undownloaded tracks to send to
	trackQueue chan metadata.Track

	// outQueue is a queue of downloaded tracks to be read from
	outQueue chan trackInfo

	// outStream is the one-stop-shop for audio data
	outStream chan []byte

	// stateChan is used to transmit state changes
	stateChan chan State

	skipChan chan bool

	// encode is an optional encoding function to run on downloaded bytes
	encode func([]byte) ([][]byte, error)
}

func NewPlayer() *Player {
	//p := &Player{}
	return &Player{
		session:    nil,
		state:      PlayingState,
		isStarted:  false,
		trackQueue: make(chan metadata.Track, 100),
		outQueue:   make(chan trackInfo, 2),
		outStream:  make(chan []byte),
		stateChan:  make(chan State),
		skipChan:   make(chan bool),
	}
}

// TODO this should be way better, but might need to fork librespot to allow it to be
func (p *Player) Login() error {
	if p.session != nil {
		return ErrPlayerAlreadyLoggedIn
	}

	var err error
	p.session, err = librespot.LoginOAuth("georgetuney", os.Getenv("SPOTIFY_ID"), os.Getenv("SPOTIFY_SECRET"))
	if err != nil {
		return fmt.Errorf("failed to initialize spotify client: %w", err)
	}

	return nil
}

func (p *Player) EncodeFunc(fn func([]byte) ([][]byte, error)) {
	p.encode = fn
}

func (p *Player) Start(ctx context.Context) {
	if p.isStarted {
		return
	}

	p.trackManager(ctx)
	p.stateManager(ctx)
	p.outStreamManager(ctx)

	p.isStarted = true
}

func (p *Player) OutStream() <-chan []byte {
	return p.outStream
}

func (p *Player) Skip() {
	p.skipChan <- true
}

//func (p *Player) Stop() {
//	for len(p.queue) > 0 {
//		<-p.queue
//	}
//	if p.state == PlayingState {
//		p.stopChan <- true
//	}
//}

// https://open.spotify.com/playlist/6yfxykP8KUzPO13V5ryCs5?si=b25ce568f1fc480e
// https://open.spotify.com/track/65LjCuIAEUX2AmWUjD9tA3?si=93c52e238e6b43f7

// SearchTrack searches for a track in spotify. The track and limit parameters are required, but artist and album are
// optional.
func (p *Player) SearchTrack(track string, artist string, album string, limit int) ([]metadata.Track, error) {
	if track == "" {
		return nil, errors.New("track must be specified")
	}

	searchResponse, err := p.session.Mercury().Search(track, limit, p.session.Country(), p.session.Username())
	if err != nil {
		return nil, err
	}

	var results []metadata.Track
	for _, resultTrack := range searchResponse.Results.Tracks.Hits {
		artistMatch := true
		albumMatch := true

		if artist != "" {
			if strings.ToLower(resultTrack.Artists[0].Name) != strings.ToLower(artist) {
				artistMatch = false
			}
		}
		if album != "" {
			if strings.ToLower(resultTrack.Album.Name) != strings.ToLower(album) {
				albumMatch = false
			}
		}
		if artistMatch && albumMatch {
			results = append(results, resultTrack)
		}
	}

	return results, nil
}

func (p *Player) QueueTrack(track metadata.Track) {
	p.queueInfo = append(p.queueInfo, track)
	p.trackQueue <- track
}

func (p *Player) GetTrackByID(trackID string) ([]byte, error) {
	track, err := p.session.Mercury().GetTrack(utils.Base62ToHex(trackID))
	if err != nil {
		fmt.Println("Error loading track: ", err)
		return nil, err
	}

	var selectedFile *Spotify.AudioFile
	for _, file := range track.GetFile() {
		if file.GetFormat() == Spotify.AudioFile_OGG_VORBIS_160 {
			selectedFile = file
		}
	}
	if selectedFile == nil {
		return nil, err
	}

	audioFile, err := p.session.Player().LoadTrack(selectedFile, track.GetGid())
	if err != nil {
		return nil, err
	}

	return ioutil.ReadAll(audioFile)
}

func (p *Player) GetTrack(track metadata.Track) ([]byte, error) {
	return p.GetTrackByID(strings.Split(track.Uri, ":")[2])
}

func (p *Player) Status() string {
	if p.currentlyPlaying.Name == "" {
		return ""
	}

	switch p.state {
	case PlayingState:
		elapsedDuration, err := time.ParseDuration(fmt.Sprintf("%fs", p.currentlyPlaying.TimeElapsed))
		if err != nil {
			return ""
		}
		lengthDuration, err := time.ParseDuration(fmt.Sprintf("%dms", p.currentlyPlaying.Duration))
		if err != nil {
			return ""
		}

		return fmt.Sprintf("%s - %s (%s/%s)", p.currentlyPlaying.Name, p.currentlyPlaying.Artists[0].Name, elapsedDuration.String(), lengthDuration.String())
		//
	case PausedState:
		//
	case StoppedState:
		//
	}

	return "unknown status"
}

func (p *Player) Queue() string {
	queueStr := ""
	for _, track := range p.queueInfo {
		queueStr += fmt.Sprintf("%s - %s\n", track.Name, track.Artists[0].Name)
	}
	return queueStr
}

func (p *Player) trackManager(ctx context.Context) {
	go func() {
		for {
			select {
			// Player should begin listening on queueChan for incoming requests
			case track := <-p.trackQueue:
				audioBytes, err := p.GetTrack(track)
				if err != nil {
					log.Error().Err(err).Str("track", track.Name).Msg("failed to get track")
					continue
				}

				encodedTrack, err := p.encode(audioBytes)
				if err != nil {
					log.Error().Err(err).Str("track", track.Name).Msg("failed to encode track")
					continue
				}

				// Tracks should be loaded off of queue chan and their data should be stored in audioBuffer
				// If audio buffer is currently full, this should block instead.
				p.outQueue <- trackInfo{Track: track, data: encodedTrack}
			case <-ctx.Done():
				return
			}
		}
	}()
}

func (p *Player) stateManager(ctx context.Context) {
	go func() {
		for {
			select {
			case state := <-p.stateChan:
				if state != p.state {

				}
			case <-ctx.Done():
				return
			}
		}
	}()
}

// outStreamManager launches a
func (p *Player) outStreamManager(ctx context.Context) {
	go func() {
		for {
			select {
			case trackData := <-p.outQueue:
				p.currentlyPlaying = TrackData{Track: trackData.Track, TimeElapsed: 0.0}
				if len(p.queueInfo) > 1 {
					p.queueInfo = p.queueInfo[1:]
				} else {
					p.queueInfo = nil
				}
			TrackLoop:
				for _, sample := range trackData.data {
					select {
					//case <-p.playChan:
					//	<-p.playChan
					case <-p.skipChan:
						break TrackLoop
					//case <-p.pauseChan:
					//	p.playChan <- false
					default:
					}
					p.outStream <- sample
					//p.outStream <- TrackData{TimeElapsed: timeElapsed, Sample: sample}
					p.currentlyPlaying.TimeElapsed += 20.0 / 1000 // 20ms
					//timeElapsed += 20.0 / 1000 // 20 ms
				}
			case <-ctx.Done():
				return
			}
		}
	}()
}

/*
player := spotify.NewPlayer()
tracks := player.SearchTrack("your graduation", "stand atlantic")
player.Queue(tracks[0])
outStream := player.OutStream()
player.Skip()
player.Pause()
player.ClearQueue()

for b := range outStream {

}


*/
