package worker

import (
	"time"

	"github.com/sauromates/leech/client"
	"github.com/sauromates/leech/internal/message"
	"github.com/sauromates/leech/internal/piece"
)

const (
	// MaxBlockSize is the largest number of bytes a request can ask for
	// (default is 16 KiB (16384 bytes)).
	MaxBlockSize int = 16384
	// MaxBacklog is the number of unfulfilled requests a client can have
	// in its pipeline.
	MaxBacklog int = 5
)

// Pipeline keeps track of piece download process while keeping backlog
// of pending requests and current download state.
type Pipeline struct {
	Client     *client.Client
	Piece      *piece.Piece
	Downloaded int
	Requested  int
	Backlog    int
}

// NewPipeline returns new pipeline with empty state (all fields set to 0).
func NewPipeline(c *client.Client, p *piece.Piece) *Pipeline {
	return &Pipeline{c, p, 0, 0, 0}
}

// Run triggers download process via pipeline. Pipeline will fill it's backlog
// with requests for piece's blocks and then read incoming messages.
func (p *Pipeline) Run() error {
	p.Client.Conn.SetDeadline(time.Now().Add(30 * time.Second))
	defer p.Client.Conn.SetDeadline(time.Time{})

	for p.Downloaded < p.Piece.Size() {
		// Send requests for blocks until the backlog is full
		if !p.Client.IsChoked {
			if err := p.requestBlocks(); err != nil {
				return err
			}
		}

		if err := p.readMessage(); err != nil {
			return err
		}
	}

	return nil
}

// requestBlocks sends block requests until the backlog is full.
func (p *Pipeline) requestBlocks() error {
	for p.hasBacklogSpace() {
		i, off, block := p.Piece.Index, p.Requested, p.blockSize()
		if err := p.Client.RequestPiece(i, off, block); err != nil {
			return err
		}

		p.Backlog++
		p.Requested += block
	}

	return nil
}

// readMessage processes responses from the connected peer.
func (p *Pipeline) readMessage() error {
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
		downloaded, err := msg.ParsePiece(p.Piece)
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
func (p *Pipeline) hasBacklogSpace() bool {
	return p.Backlog < MaxBacklog && p.Requested < p.Piece.Size()
}

// blockSize returns either a constant size of a block to request or (in case
// of the last block) the calculated last block size
func (p *Pipeline) blockSize() int {
	blockSize := p.Piece.Size() - p.Requested
	if blockSize < MaxBlockSize {
		return blockSize
	}

	return MaxBlockSize
}
