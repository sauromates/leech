// Package piece provides a type for downloadable torrents' parts - pieces.
//
// For the sake of convenience pieces implement [io.Writer], [io.WriterAt]
// and [io.Reader], [io.ReaderAt] since they need to be downloaded and copied.
package piece

import (
	"bytes"
	"crypto/sha1"
	"errors"
	"fmt"
	"io"

	"github.com/sauromates/leech/internal/bthash"
)

var (
	// ErrInvalidHashSum is returned when downloaded content's SHA-1 hash sum
	// differs from piece's hash.
	ErrInvalidHashSum error = errors.New("integrity check failed")
)

// Piece is a downloadable part of torrents. Every torrent consists of pieces
// of fixed sized which are retrieved from peers.
type Piece struct {
	Index   int         // Index defines piece's position in the torrent
	Hash    bthash.Hash // Hash is [20]byte identifier of piece contents
	Offset  int64       // Absolute start position in the torrent
	End     int64       // Absolute end position in the torrent
	Content []byte      // Content holds piece's actual data
	section []byte      // Internal buffer of writable contents
}

// New creates new piece with empty content of given size.
func New(index int, hash bthash.Hash, offset, end int64) *Piece {
	return &Piece{
		Index:   index,
		Hash:    hash,
		Offset:  offset,
		End:     end,
		Content: make([]byte, end-offset),
	}
}

// Bounds calculate absolute start and end position of a piece within
// a torrent based on piece's index, torrent's default piece length and
// total torrent's length.
func Bounds(index, length, maxLength int) (offset, end int) {
	offset = index * length
	end = min(offset+length, maxLength)

	return offset, end
}

// Size returns piece length calculated as difference between absolute end
// position and absolute start position within the torrent.
func (p *Piece) Size() int { return max(int(p.End-p.Offset), len(p.Content)) }

// VerifyHash checks that piece hash matches SHA-1 sum of its content.
func (p *Piece) VerifyHash() error {
	hash := sha1.Sum(p.Content)
	if !bytes.Equal(hash[:], p.Hash[:]) {
		return ErrInvalidHashSum
	}

	return nil
}

// Section puts slice of piece's content into internal writable buffer.
func (p *Piece) Section(off, end int64) *Piece {
	p.section = p.Content[off:end]

	return p
}

// WriteTo copies writable piece content to specified writer directly.
//
// If writable section wasn't set beforehand via [Piece.Section] method it
// writes the whole piece content.
func (p *Piece) WriteTo(w io.Writer) (n int64, err error) {
	// When writable section is empty we write the whole piece content
	if len(p.section) == 0 {
		p.section = p.Content
	}

	written, err := w.Write(p.section)

	// Truncate writable section after it was written
	p.section = []byte{}

	return int64(written), err
}

// Write copies given slice of bytes into the [Piece.Content] field. Given
// slice length MUST be less than or equal to piece size.
func (p *Piece) Write(b []byte) (n int, err error) {
	if len(b) > p.Size() {
		return n, fmt.Errorf("content is larger than piece size: %d", len(b))
	}

	return copy(p.Content[:], b), err
}

// Write copies provided bytes to [Piece.Content] field starting at the given
// offset. Offset must be within [Piece.Length].
func (p *Piece) WriteAt(b []byte, off int64) (n int, err error) {
	if int(off) > p.Size() {
		return n, fmt.Errorf("offset %d is outside piece bounds", off)
	}

	return copy(p.Content[off:], b), err
}

// Read calls underlying [*bytes.Reader] to read piece contents.
func (p *Piece) Read(b []byte) (n int, err error) {
	return p.reader().Read(b)
}

// ReadAt calls underlying [*bytes.Reader] to read piece contents from offset.
func (p *Piece) ReadAt(b []byte, off int64) (n int, err error) {
	return p.reader().ReadAt(b, off)
}

// reader returns new [*bytes.Reader] instance to be used later for [io.Reader]
// and [io.ReaderAt] implementations.
func (p *Piece) reader() *bytes.Reader { return bytes.NewReader(p.Content) }
