package apollo

import "io"

type Transcoder interface {
	Start(io.Reader) error
	Cancel()
	OutBytes() <-chan []byte
	Errors() <-chan error
}
