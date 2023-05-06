package ogg

import (
	"bytes"
	"io"
)

type Decoder struct {
	r           io.Reader
	currentPage *Page
	lastSegment int
}

func NewDecoder() *Decoder {
	return &Decoder{
		//r:           r,
		currentPage: nil,
		lastSegment: 0,
	}
}

func (d *Decoder) Open(r io.Reader) error {
	d.r = r
	return nil
}

// Next returns the next packet of the stream. Returns io.EOF if all packets have already been read.
func (d *Decoder) Next() ([]byte, error) {
	var buf bytes.Buffer

	readComplete := false
	for !readComplete {
		if d.currentPage == nil || d.lastSegment >= len(d.currentPage.SegmentTable) {
			page, err := ReadPage(d.r)
			if err != nil {
				return nil, err
			}

			if len(page.SegmentTable) == 0 {
				continue
			}

			d.currentPage = &page
			d.lastSegment = 0
		}

		if _, err := io.CopyN(&buf, d.r, int64(d.currentPage.SegmentTable[d.lastSegment])); err != nil {
			return nil, err
		}

		readComplete = d.currentPage.SegmentTable[d.lastSegment] < 255
		d.lastSegment++
	}

	return buf.Bytes(), nil
}

// Decode decodes the next packet into p. The number of bytes written and any errors encountered are returned. If p is
// smaller than the packet read, io.ErrShortBuffer will be returned.
func (d *Decoder) Read(p []byte) (int, error) {
	pos := 0
	bytesToRead := 0

	for {
		if d.currentPage == nil || d.lastSegment >= len(d.currentPage.SegmentTable) {
			page, err := ReadPage(d.r)
			if err != nil {
				return 0, err
			}

			if len(page.SegmentTable) == 0 {
				continue
			}

			d.currentPage = &page
			d.lastSegment = 0
		}

		// Calculate how many bytes to be read
		for ; d.lastSegment < len(d.currentPage.SegmentTable); d.lastSegment++ {
			bytesToRead += int(d.currentPage.SegmentTable[d.lastSegment])
			if d.currentPage.SegmentTable[d.lastSegment] < 255 {
				d.lastSegment++
				break
			}
		}

		if len(p) < bytesToRead {
			return 0, io.ErrShortBuffer
		}

		// If we hit the end of the segment table with more data on the next page
		if d.lastSegment == len(d.currentPage.SegmentTable) && d.currentPage.SegmentTable[d.lastSegment-1] == 255 {
			n, err := d.r.Read(p[pos:bytesToRead])
			if err != nil {
				return bytesToRead, err
			}

			pos += n
			continue
		}

		return d.r.Read(p[pos:bytesToRead])
	}
}

func (d *Decoder) ReadAll() ([][]byte, error) {
	var all [][]byte
	buf := make([]byte, MaxPageSize)

	for {
		n, err := d.Read(buf)
		if err != nil {
			return all, err
		}
		packet := make([]byte, n)
		copy(packet, buf[:n])

		all = append(all, packet)
	}
}

func (d *Decoder) Close() error {
	d.r = nil
	d.currentPage = nil
	d.lastSegment = 0
	return nil
}
