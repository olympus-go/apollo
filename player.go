package apollo

import (
	"context"
	"io"
	"log/slog"
	"math/rand"
	"time"

	"github.com/eolso/threadsafe"
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
	config PlayerConfig
	codec  Codec

	cursor int
	queue  *threadsafe.Slice[Playable]

	currentState PlayerState
	stateChan    chan PlayerState

	outChan    chan []byte
	bytesSent  int
	playCancel context.CancelFunc

	logger *slog.Logger
}

// NewPlayer creates a new player instance and starts listening for events. If no logging is desired, nil can be passed
// in for h.
func NewPlayer(config PlayerConfig, h slog.Handler) *Player {
	if h == nil {
		h = nopLogHandler{}
	}

	p := Player{
		config:       config,
		codec:        &NopCodec{},
		cursor:       0,
		queue:        &threadsafe.Slice[Playable]{},
		currentState: IdleState,
		stateChan:    make(chan PlayerState),
		outChan:      make(chan []byte),
		logger:       slog.New(h),
	}

	go p.stateListener()

	return &p
}

func (p *Player) WithCodec(c Codec) *Player {
	p.codec = c
	return p
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
	p.logger.Info("enqueued "+playable.Type(), slog.Any("playable", nameArtistAlbumType(playable)))
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

// Get returns the Playable at position i. Returns nil when i is invalid.
func (p *Player) Get(i int) Playable {
	pl, _ := p.queue.SafeGet(i)
	return pl
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
	if p.playCancel != nil {
		p.playCancel()
	}
	p.queue.Empty()
	p.cursor = 0
	p.bytesSent = 0
}

func (p *Player) Shuffle(all bool) {
	start := 0
	end := p.queue.Len()

	if !all {
		start = p.cursor
	}

	shuffledQueue := threadsafe.Slice[Playable]{}

	for i := start; i < p.queue.Len(); i++ {
		shuffledQueue.Append(p.queue.Get(i))
	}

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	rng.Shuffle(end-start, func(i, j int) {
		shuffledQueue.Data[i], shuffledQueue.Data[j] = shuffledQueue.Data[j], shuffledQueue.Data[i]
	})

	newQueue := threadsafe.Slice[Playable]{}

	for i := 0; i < start; i++ {
		if !all {
			newQueue.Append(p.queue.Get(i))
		} else {
			newQueue.Append(shuffledQueue.Get(i))
		}
	}

	for i := start; i < end; i++ {
		newQueue.Append(shuffledQueue.Get(i - start))
	}

	p.queue = &newQueue
}

func (p *Player) NowPlaying() (Playable, bool) {
	if p.currentState == PlayState || p.currentState == PauseState {
		return p.queue.SafeGet(p.cursor - 1)
	}

	return nil, false
}

func (p *Player) Cursor() int {
	return p.cursor
}

func (p *Player) State() PlayerState {
	return p.currentState
}

func (p *Player) Out() <-chan []byte {
	return p.outChan
}

func (p *Player) BytesSent() int {
	// TODO the player should probably track time itself, but this will be a quicker win.
	return p.bytesSent
}

// stateListener handles all the state change requests. This routine also launches the playable listener and establishes
// a channel to communicate with it.
func (p *Player) stateListener() {
	logger := p.logger.With(slog.String("goroutine", "stateListener()"))
	processChan, playChan := p.playableListener()

	for {
		select {
		case state := <-p.stateChan:
			logger.Debug("received request for state change",
				slog.String("current", p.currentState.String()),
				slog.String("requested", state.String()),
			)

			switch state {
			case IdleState:
				p.currentState = IdleState
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

func (p *Player) playableListener() (chan<- PlayerState, chan<- Playable) {
	logger := p.logger.With(slog.String("goroutine", "playableListener()"))
	stateChan := make(chan PlayerState)
	playChan := make(chan Playable)

	go func() {
		var playerCtx context.Context
		buf := make([]byte, p.config.PacketBuffer)

		for {
			playerCtx, p.playCancel = context.WithCancel(context.Background())

			select {
			case s := <-stateChan:
				// Nothing to do if not currently playing, but we don't want to have the channel backed up when idle
				logger.Debug("discarded state change request", slog.String("requested", s.String()))
			case playable := <-playChan:
				p.bytesSent = 0

				logger.Info("downloading "+playable.Type(), slog.Any("playable", nameArtistAlbumType(playable)))

				r, err := playable.Download()
				if err != nil {
					logger.Error("failed to download as "+playable.Type(),
						slog.String("error", err.Error()),
						slog.Any("playable", nameArtistAlbumType(playable)),
					)
					continue
				}

				err = p.codec.Open(r)
				if err != nil {
					slog.Error("failed to open as "+playable.Type(),
						slog.String("error", err.Error()),
						slog.Any("playable", nameArtistAlbumType(playable)),
					)
					continue
				}

				func() {
					for {
						select {
						case <-playerCtx.Done():
							logger.Debug("player context closed")
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
											logger.Debug("skipping "+playable.Type(),
												slog.Any("playable", nameArtistAlbumType(playable)),
											)
											// In the case of a PlayState being received, we just need to stop blocking
											// here. But when a NextState is received, we need to stop blocking and
											// signal parent loop to cancel.
											p.playCancel()
											return
										}
									}
								}()
							case NextState:
								logger.Debug("skipping "+playable.Type(),
									slog.Any("playable", nameArtistAlbumType(playable)),
								)
								return
							}
						default:
							n, err := p.codec.Read(buf)
							if err != nil && err == io.EOF {
								logger.Info("finished playing "+playable.Type(),
									slog.Any("playable", nameArtistAlbumType(playable)),
								)
								return
							} else if err != nil {
								slog.Error("error reading "+playable.Type(),
									slog.String("error", err.Error()),
									slog.Any("playable", nameArtistAlbumType(playable)),
								)
								return
							}

							out := make([]byte, n)
							copy(out, buf[:n])

							select {
							case <-playerCtx.Done():
								return
							case p.outChan <- out:
								p.bytesSent++
							}
						}
					}
				}()

				if err = p.codec.Close(); err != nil {
					logger.Error("failed closing codec", slog.String("error", err.Error()))
				}

				if err = r.Close(); err != nil {
					logger.Error("failed closing "+playable.Type(),
						slog.String("error", err.Error()),
						slog.Any("playable", nameArtistAlbumType(playable)),
					)
				}

				// Attempt to play the next in queue
				p.stateChan <- IdleState
				p.stateChan <- PlayState
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

func (p *Player) idle() {
	go func() {
		p.stateChan <- IdleState
	}()
}
