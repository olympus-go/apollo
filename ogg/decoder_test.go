package ogg_test

import (
	"bytes"
	"errors"
	"io"
	"math/rand"
	"testing"
	"time"

	"github.com/olympus-go/apollo/ogg"
)

func TestDecoder_Read(t *testing.T) {
	type test struct {
		r   io.Reader
		buf []byte
		err error
	}

	tests := map[string]test{
		"nil_reader":       {nil, nil, ogg.ErrInvalid},
		"invalid_data":     {bytes.NewReader(make([]byte, 100)), make([]byte, 100), ogg.ErrInvalid},
		"short_buffer":     {genOgg(10, 500), make([]byte, 100), io.ErrShortBuffer},
		"valid_data_empty": {genOgg(0, 100), make([]byte, 100), io.EOF},
		"valid_data_many":  {genOgg(10, 1024), make([]byte, 1024), nil},
	}

	for name, tst := range tests {
		t.Run(name, func(t *testing.T) {
			d := ogg.NewDecoder(tst.r)
			_, err := d.Read(tst.buf)

			if tst.err != err {
				if tst.err == nil {
					tst.err = errors.New("nil")
				}
				if err == nil {
					err = errors.New("nil")
				}
				t.Fatalf("expected %q error; got %q", tst.err.Error(), err.Error())
			}
		})
	}
}

func BenchmarkDecoder_Read(b *testing.B) {
	type test struct {
		r   io.Reader
		buf []byte
	}

	tests := map[string]test{
		"right_sized_100":   {genOgg(100, 1024), make([]byte, 1024)},
		"right_sized_1000":  {genOgg(1000, 1024), make([]byte, 1024)},
		"right_sized_10000": {genOgg(10000, 1024), make([]byte, 1024)},
		"over_sized_100":    {genOgg(100, 1024), make([]byte, ogg.MaxPageSize)},
		"over_sized_1000":   {genOgg(1000, 1024), make([]byte, ogg.MaxPageSize)},
		"over_sized_10000":  {genOgg(10000, 1024), make([]byte, ogg.MaxPageSize)},
	}

	for name, tst := range tests {
		d := ogg.NewDecoder(tst.r)

		b.Run(name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				for {
					_, err := d.Read(tst.buf)
					if err != nil && err == io.EOF {
						break
					} else if err != nil {
						b.Fatal(err)
					}
				}
			}
		})
	}
}

func BenchmarkDecoder_Next(b *testing.B) {
	tests := map[string]int{
		"100":   100,
		"1000":  1000,
		"10000": 10000,
	}

	for name, test := range tests {
		r := genOgg(test, 1024)
		d := ogg.NewDecoder(r)
		b.Run(name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				for {
					buf, err := d.Next()
					_ = buf
					if err != nil && err == io.EOF {
						break
					} else if err != nil {
						b.Fatal(err)
					}
				}
			}
		})
	}
}

func TestDecoder_ReadAll(t *testing.T) {
	type test struct {
		r   io.Reader
		err error
	}

	tests := map[string]test{
		"nil_reader":       {nil, ogg.ErrInvalid},
		"invalid_data":     {bytes.NewReader(make([]byte, 100)), ogg.ErrInvalid},
		"valid_data_empty": {genOgg(0, 2048), io.EOF},
		"valid_data_many":  {genOgg(10, 2048), io.EOF},
	}

	for name, tst := range tests {
		t.Run(name, func(t *testing.T) {
			d := ogg.NewDecoder(tst.r)
			_, err := d.ReadAll()

			if tst.err != err {
				if tst.err == nil {
					tst.err = errors.New("nil")
				}
				if err == nil {
					err = errors.New("nil")
				}
				t.Fatalf("expected %q error; got %q", tst.err.Error(), err.Error())
			}
		})
	}
}

// genOgg creates a reader to a dummy ogg file with nPages containing packets of size packetSize bytes
func genOgg(nPages int, packetSize int) io.Reader {
	var buf bytes.Buffer
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))

	currentSize := 0
	for i := 0; i < nPages; i++ {
		pageLen := rnd.Int()%10 + 1

		header := ogg.PageHeader{
			CapturePattern:         ogg.CapturePattern,
			StreamStructureVersion: 0,
			HeaderTypeFlag:         0,
			GranulePosition:        0,
			BitstreamSerialNumber:  0,
			PageSequenceNumber:     uint32(i),
			CRCChecksum:            0,
			NumberPageSegments:     uint8(pageLen),
		}

		segmentTable := make([]byte, pageLen)

		for segmentIndex := range segmentTable {
			if currentSize == packetSize || currentSize+255 == packetSize {
				segmentTable[segmentIndex] = 0
				currentSize = 0
			} else if currentSize+255 < packetSize {
				segmentTable[segmentIndex] = 255
				currentSize += 255
			} else if currentSize+255 >= packetSize {
				segmentTable[segmentIndex] = byte(packetSize - currentSize)
				currentSize = 0
			}
		}

		page := ogg.Page{
			Header:       header,
			SegmentTable: segmentTable,
		}

		// Ensure that the last page doesn't end with a dangling packet.
		if i == nPages-1 {
			page.SegmentTable[page.Header.NumberPageSegments-1] = 0
		}

		buf.Write(page.Serialize())

		for _, table := range segmentTable {
			buf.Write(make([]byte, table))
		}
	}

	return bytes.NewReader(buf.Bytes())
}
