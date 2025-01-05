package worker

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	_ "io"

	"github.com/sauromates/leech/client"
	"github.com/sauromates/leech/internal/message"
	"github.com/sauromates/leech/internal/utils"
)

const (
	// MaxBlockSize is the largest number of bytes a request can ask for
	// (default is 2048 Kb (16384 bytes))
	MaxBlockSize int = 16384
	// MaxBacklog is the number of unfulfilled requests a client can have
	// in its pipeline
	MaxBacklog int = 5
)

// Piece represents downloadable piece
type Piece struct {
	Index  int
	Hash   utils.BTString
	Length int
}

// PieceContent holds contents of downloaded piece
//
// It also implements [io.Reader] and [io.ReaderAt] to allow results usage
// in I/O operations
type PieceContent struct {
	Index   int
	Content []byte
}

// Pipeline keeps track of piece download process while keeping backlog
// of pending requests and current download state
type Pipeline struct {
	Index      int
	Client     *client.Client
	Content    []byte
	Downloaded int
	Requested  int
	Backlog    int
}

// ReadMessage processes responses from the connected peer
func (p *Pipeline) ReadMessage() error {
	msg, err := p.Client.Read() // This call blocks
	if err != nil {
		return err
	}

	if msg == nil {
		return nil
	}

	switch msg.ID {
	case message.Unchoke:
		p.Client.IsChoked = false
	case message.Choke:
		p.Client.IsChoked = true
	case message.Have:
		index, err := msg.ParseHave()
		if err != nil {
			return err
		}

		p.Client.BitField.SetPiece(index)
	case message.Piece:
		downloaded, err := msg.ParsePiece(p.Index, p.Content)
		if err != nil {
			return err
		}

		p.Downloaded += downloaded
		p.Backlog--
	}

	return nil
}

// hasBacklogSpace determines whether the worker can accumulate more requests
// for given task
func (p *Pipeline) hasBacklogSpace(pieceLength int) bool {
	return p.Backlog < MaxBacklog && p.Requested < pieceLength
}

// blockSize returns either a constant size of a block to request or (in case
// of the last block) the calculated last block size
func (p *Pipeline) blockSize(pieceLength int) int {
	blockSize := pieceLength - p.Requested
	if blockSize < MaxBlockSize {
		return blockSize
	}

	return MaxBlockSize
}

// verifyHashSum compares sha1 hash sums of a piece and downloaded content
func (p *Piece) verifyHashSum(content []byte) error {
	hash := sha1.Sum(content)
	if !bytes.Equal(hash[:], p.Hash[:]) {
		return fmt.Errorf("piece %d failed integrity check", p.Index)
	}

	return nil
}

// reader returns new [*bytes.Reader] instance to be used later for [io.Reader]
// and [io.ReaderAt] implementations
func (c *PieceContent) reader() *bytes.Reader {
	return bytes.NewReader(c.Content)
}

// Read calls underlying [*bytes.Reader] to read piece contents
func (c *PieceContent) Read(b []byte) (n int, err error) {
	return c.reader().Read(b)
}

// ReadAt calls underlying [*bytes.Reader] to read piece contents from offset
func (c *PieceContent) ReadAt(b []byte, off int64) (n int, err error) {
	return c.reader().ReadAt(b, off)
}
