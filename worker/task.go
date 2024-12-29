package worker

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"log"

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
	MaxBacklog int = 50
)

// Task represents downloadable piece
type Task struct {
	Index  int
	Hash   utils.BTString
	Length int
}

// TaskResult holds contents of downloaded piece
type TaskResult struct {
	Index   int
	Content []byte
}

// TaskProgress keeps track of piece download process
type TaskProgress struct {
	Index      int
	Client     *client.Client
	Content    []byte
	Downloaded int
	Requested  int
	Backlog    int
}

// ReadMessage processes responses from the connected peer
func (state *TaskProgress) ReadMessage() error {
	msg, err := state.Client.Read() // This call blocks
	if err != nil {
		return err
	}

	if msg == nil {
		return nil
	}

	log.Printf("[INFO] Received %v from %s", msg, state.Client.Peer.String())

	switch msg.ID {
	case message.Unchoke:
		state.Client.IsChoked = false
	case message.Choke:
		state.Client.IsChoked = true
	case message.Have:
		index, err := msg.ParseHave()
		if err != nil {
			return err
		}

		state.Client.BitField.SetPiece(index)
	case message.Piece:
		downloaded, err := msg.ParsePiece(state.Index, state.Content)
		if err != nil {
			return err
		}

		state.Downloaded += downloaded
		state.Backlog--
	}

	return nil
}

// hasBacklogSpace determines whether the worker can accumulate more requests
// for given task
func (state *TaskProgress) hasBacklogSpace(pieceLength int) bool {
	return state.Backlog < MaxBacklog && state.Requested < pieceLength
}

// blockSize returns either a constant size of a block to request or (in case
// of the last block) the calculated last block size
func (state *TaskProgress) blockSize(pieceLength int) int {
	blockSize := pieceLength - state.Requested
	if blockSize < MaxBlockSize {
		return blockSize
	}

	return MaxBlockSize
}

// verifyHashSum compares sha1 hash sums of a piece and downloaded content
func (piece *Task) verifyHashSum(content []byte) error {
	hash := sha1.Sum(content)
	if !bytes.Equal(hash[:], piece.Hash[:]) {
		return fmt.Errorf("piece %d failed integrity check", piece.Index)
	}

	return nil
}
