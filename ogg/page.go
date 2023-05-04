package ogg

import (
	"bytes"
	"encoding/binary"
	"io"
)

// Implementation spec: https://www.xiph.org/ogg/doc/rfc3533.txt

// TODO actually do CRC checks hehe
const generatorPolynomial = `0x04c11db7`
const MaxPageSize = 65307

// ByteOrder is the byte order used by ogg containers.
var ByteOrder = binary.LittleEndian

// CapturePattern is the 4 byte field used to denote the beginning of a ogg page header.
var CapturePattern = [4]byte{'O', 'g', 'g', 'S'}

// PageHeader represents the header metadata of an OGG page. Each header should be exactly 27 bytes.
type PageHeader struct {
	// Byte field that signifies the beginning of a page. This should always be 'OggS'.
	CapturePattern [4]byte

	// The version of Ogg used in the stream.
	StreamStructureVersion byte

	// The bits of this byte are used to determine properties of the page.
	//   0x01:
	//   - 0: page contains a fresh packet
	//   - 1: page contains data of a packet continued from the previous page
	//   0x02:
	//   - 0: this page is not a first page
	//   - 1: this is the first page of a logical bitstream (bos)
	//   0x04:
	//   - 0: this page is not a last page
	//   - 1: this is the last page of a logical bitstream (eos)
	HeaderTypeFlag byte

	// 8 byte field containing position information. Meaning of value is dependent on the codec being used.
	GranulePosition int64

	// 4 byte unique serial number identifying the bitstream.
	BitstreamSerialNumber uint32

	// 4 byte field containing the sequence number of the page. Used to detect page loss.
	PageSequenceNumber uint32

	// 4 byte field containing a 32-bit CRC checksum of the page.
	CRCChecksum uint32

	// 1 Byte giving the number of segment entries encoded in the segment table.
	NumberPageSegments uint8
}

type Page struct {
	Header PageHeader

	// Byte slice of size Header.NumberPageSegments containing the lacing values of all segments in this page.
	SegmentTable []byte
}

func ReadHeader(r io.Reader) (PageHeader, error) {
	var header PageHeader

	if r == nil {
		return header, ErrInvalid
	}

	err := binary.Read(r, ByteOrder, &header)
	if err != nil {
		return header, err
	}

	if header.CapturePattern != CapturePattern {
		return header, ErrInvalid
	}

	return header, err
}

func ReadPage(r io.Reader) (page Page, err error) {
	page.Header, err = ReadHeader(r)
	if err != nil {
		return
	}

	page.SegmentTable = make([]byte, page.Header.NumberPageSegments)
	_, err = io.ReadFull(r, page.SegmentTable)

	return
}

func (p PageHeader) Serialize() []byte {
	var buf bytes.Buffer
	buf.Grow(27)

	_ = binary.Write(&buf, ByteOrder, p.CapturePattern)
	_ = binary.Write(&buf, ByteOrder, p.StreamStructureVersion)
	_ = binary.Write(&buf, ByteOrder, p.HeaderTypeFlag)
	_ = binary.Write(&buf, ByteOrder, p.GranulePosition)
	_ = binary.Write(&buf, ByteOrder, p.BitstreamSerialNumber)
	_ = binary.Write(&buf, ByteOrder, p.PageSequenceNumber)
	_ = binary.Write(&buf, ByteOrder, p.CRCChecksum)
	_ = binary.Write(&buf, ByteOrder, p.NumberPageSegments)

	return buf.Bytes()
}

func (p Page) Serialize() []byte {
	var buf bytes.Buffer

	buf.Write(p.Header.Serialize())
	buf.Write(p.SegmentTable)

	return buf.Bytes()
}
