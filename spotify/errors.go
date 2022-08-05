package spotify

import "errors"

var ErrPlayerAlreadyLoggedIn = errors.New("spotify: player already logged in")
var ErrEmptySearchResponse = errors.New("spotify: search yielded no results")
