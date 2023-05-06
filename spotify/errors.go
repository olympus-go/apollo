package spotify

import (
	"errors"
)

var ErrPlayerQueueFull = errors.New("player queue is full")
var ErrPlayerCommandTimeout = errors.New("player command timed out")
var ErrTokenNotFound = errors.New("spotify: auth token not found")
var ErrPlayerAlreadyLoggedIn = errors.New("spotify: player already logged in")
var ErrEmptySearchResponse = errors.New("spotify: search yielded no results")
