package spotify

// Cache should represent a FILO-able structure
type Cache interface {
	Push([]byte)
	PopBack() []byte
}
