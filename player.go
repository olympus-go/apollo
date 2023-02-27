package apollo

import (
	"context"

	"github.com/eolso/threadsafe"
	"github.com/olympus-go/apollo/ffmpeg"
	"github.com/rs/zerolog"
)

type PlayerState int

func (p PlayerState) String() string {
	return []string{"Idle", "Play", "Pause", "Skip", "Previous"}[p]
}

const (
	// IdleState represents a paused state with nothing left in queue. This can't be specifically requested.
	IdleState PlayerState = iota
	// PlayState is when the player is actively outputting audio data.
	PlayState
	// PauseState is when the player is not currently outputting audio data, but there are things enqueued.
	PauseState
	// NextState is set briefly when the player is actively incrementing the queue cursor.
	NextState
	// PreviousState is set briefly when the player is actively decrementing the queue cursor.
	PreviousState
)

type Player struct {
	config       PlayerConfig
	transcoder   Transcoder
	cursor       int
	queue        threadsafe.Slice[Playable]
	currentState PlayerState
	stateChan    chan PlayerState
	outChan      chan []byte
	bytesSent    int
	logger       zerolog.Logger
}

// NewPlayer creates a new player instance and starts listening for events. If no logging is desired, a zerolog.Nop()
// should be used.
func NewPlayer(ctx context.Context, config PlayerConfig, log zerolog.Logger) *Player {
	p := Player{
		config:       config,
		transcoder:   ffmpeg.NewTranscoder(),
		cursor:       0,
		currentState: IdleState,
		stateChan:    make(chan PlayerState),
		outChan:      make(chan []byte),
		logger:       log.With().Str("package", "apollo").Logger(),
	}

	go p.stateListener(ctx)

	return &p
}

func (p *Player) Play() {
	go func() {
		p.stateChan <- PlayState
	}()
}

func (p *Player) Pause() {
	go func() {
		p.stateChan <- PauseState
	}()
}

func (p *Player) Enqueue(playable Playable) {
	p.queue.Append(playable)
	p.logger.Info().
		Interface("playable", nameArtistAlbumType(playable)).
		Msgf("enqueued %s", playable.Type())
}

func (p *Player) Next() {
	go func() {
		p.stateChan <- NextState
	}()
}

func (p *Player) Previous() {
	go func() {
		p.stateChan <- PreviousState
	}()
}

func (p *Player) Insert(i int, playable Playable) {
	if i < 0 {
		return
	}

	if i >= p.queue.Len() {
		p.queue.Append(playable)
	} else {
		p.queue.SafeInsert(i, playable)
	}
}

func (p *Player) Remove(i int) {
	if i < 0 || i >= p.queue.Len() {
		return
	}

	p.queue.SafeDelete(i)
}

func (p *Player) List(all bool) []Playable {
	if all {
		return p.queue.GetAll()
	}

	return p.queue.GetAll()[p.cursor:]
}

func (p *Player) Empty() {
	p.queue.Empty()
	p.transcoder.Cancel()
	p.Next()
}

func (p *Player) NowPlaying() (Playable, bool) {
	if p.currentState == PlayState || p.currentState == PauseState {
		return p.queue.SafeGet(p.cursor - 1)
	}

	return nil, false
}

func (p *Player) State() PlayerState {
	return p.currentState
}

func (p *Player) Out() <-chan []byte {
	return p.outChan
}

// TODO the player should probably track time itself, but this will be a quicker win.
func (p *Player) BytesSent() int {
	return p.bytesSent
}

func (p *Player) stateListener(ctx context.Context) {
	logger := p.logger.With().Str("goroutine", "stateListener()").Logger()
	processChan, playChan := p.playableListener(ctx)

	for {
		select {
		case <-ctx.Done():
			return
		case state := <-p.stateChan:
			logger.Debug().
				Str("current", p.currentState.String()).
				Str("requested", state.String()).
				Msg("received request for state change")
			switch state {
			case PlayState:
				if p.currentState == IdleState {
					if playable, ok := p.queue.SafeGet(p.cursor); ok {
						playChan <- playable
						p.moveCursor(1)
						p.currentState = PlayState
					}
				} else if p.currentState == PauseState {
					processChan <- PlayState
					p.currentState = PlayState
				}
			case PauseState:
				if p.currentState == PlayState {
					processChan <- PauseState
					p.currentState = PauseState
				}
			case NextState:
				if p.currentState == PlayState || p.currentState == PauseState {
					p.currentState = NextState
					processChan <- NextState
					p.currentState = IdleState
					p.Play()
				}
			case PreviousState:
				if p.queue.Len() > 0 {
					initialState := p.currentState
					p.currentState = PreviousState
					// Previous is the same as skipping in the sense that we need to cancel what is currently playing.
					// The only difference is that we need to move the cursor back before resuming playback.
					processChan <- NextState
					// When at the end we only need to go back one (idle -> previous song). If we are currently playing
					// or paused we need to go back two (playing -> beginning of song -> previous song).
					if initialState == IdleState {
						p.moveCursor(-1)
					} else if initialState == PauseState || initialState == PlayState {
						p.moveCursor(-2)
					}

					p.currentState = IdleState
					p.Play()
				}
			}
		}
	}
}

func (p *Player) playableListener(ctx context.Context) (chan<- PlayerState, chan<- Playable) {
	logger := p.logger.With().Str("goroutine", "playableListener()").Logger()
	stateChan := make(chan PlayerState)
	playChan := make(chan Playable)

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case s := <-stateChan:
				// Nothing to do if not currently playing, but we don't want to have the channel backed up when idle
				logger.Debug().Str("requested", s.String()).Msg("discarded state change request")
			case playable := <-playChan:
				p.bytesSent = 0

				logger.Info().
					Interface("playable", nameArtistAlbumType(playable)).
					Msgf("downloading %s", playable.Type())

				r, err := playable.Download()
				if err != nil {
					logger.Error().
						Err(err).
						Interface("playable", nameArtistAlbumType(playable)).
						Msgf("failed to download %s", playable.Type())
					continue
				}

				if err = p.transcoder.Start(r); err != nil {
					logger.Error().
						Err(err).
						Interface("playable", nameArtistAlbumType(playable)).
						Msgf("failed to start transcoder for %s", playable.Type())
					continue
				}

				playerCtx, playerCancel := context.WithCancel(ctx)
				func() {
					for b := range p.transcoder.OutBytes() {
						select {
						case <-playerCtx.Done():
							return
						case state := <-stateChan:
							switch state {
							case PauseState:
								func() {
									// Start blocking until we receive a Play or Skip state request
									for s := range stateChan {
										if s == PlayState {
											return
										} else if s == NextState {
											logger.Debug().
												Interface("playable", nameArtistAlbumType(playable)).
												Msgf("skipping %s", playable.Type())
											playerCancel()
											return
										}
									}
								}()
							case NextState:
								logger.Debug().
									Interface("playable", nameArtistAlbumType(playable)).
									Msgf("skipping %s", playable.Type())
								return
							}
						default:
							p.outChan <- b
							p.bytesSent++
						}
					}

					if err, ok := <-p.transcoder.Errors(); ok && err != nil {
						logger.Error().
							Err(err).
							Interface("playable", nameArtistAlbumType(playable)).
							Msgf("error while transcoding %s", playable.Type())
						return
					}

					logger.Info().
						Interface("playable", nameArtistAlbumType(playable)).
						Msgf("finished playing %s", playable.Type())
				}()

				p.transcoder.Cancel()

				// Attempt to play the next in queue
				p.currentState = IdleState
				p.Play()
			}
		}
	}()

	return stateChan, playChan
}

// moveCursor moves the cursor the by the specified amount and then checks that it is still in the accepted bounds
// [0, len(queue)]. If it is out of bounds, it sets the cursor to the nearest acceptable value.
func (p *Player) moveCursor(i int) {
	tempCursor := p.cursor
	tempCursor += i
	if tempCursor < 0 {
		tempCursor = 0
	} else if tempCursor > p.queue.Len() {
		tempCursor = p.queue.Len()
	}

	p.cursor = tempCursor
}
