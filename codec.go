package apollo

import (
	"io"
)

// Codec represents any encoder or decoder that can Read into a stream of bytes.
type Codec interface {
	Open(io.Reader) error
	io.ReadCloser
}

type NopCodec struct {
	r io.Reader
}

func (n *NopCodec) Open(r io.Reader) error {
	n.r = r
	return nil
}

func (n *NopCodec) Read(p []byte) (int, error) {
	return n.r.Read(p)
}

func (n *NopCodec) Close() error {
	return nil
}
