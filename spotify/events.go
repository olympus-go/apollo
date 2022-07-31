package spotify

type PlayerEvent int

const (
	PlayerEventPlay PlayerEvent = iota
	PlayerEventPause
	PlayerEventStop
	PlayerEventNext
	PlayerEventPrevious
	PlayerEventUnknown
)

func (p PlayerEvent) String() string {
	if p < 0 || p > PlayerEventUnknown {
		return "Unknown"
	}
	return []string{"Play", "Pause", "Stop", "Next", "Previous", "Unknown"}[p]
}
