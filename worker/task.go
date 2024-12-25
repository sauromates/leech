package worker

import (
	"bytes"
	"crypto/sha1"
	"fmt"

	"github.com/sauromates/leech/client"
	"github.com/sauromates/leech/internal/message"
	"github.com/sauromates/leech/internal/utils"
)

const (
	// MaxBlockSize is the largest number of bytes a request can ask for (default is 2048 Kb (16384 bytes))
	MaxBlockSize int = 16384
	// MaxBacklog is the number of unfulfilled requests a client can have in its pipeline
	MaxBacklog int = 5
)

type TaskItem struct {
	Index  int
	Hash   utils.BTString
	Length int
}

type TaskResult struct {
	Index   int
	Content []byte
}

type TaskProgress struct {
	Index      int
	Client     *client.Client
	Content    []byte
	Downloaded int
	Requested  int
	Backlog    int
}

func (state *TaskProgress) ReadMessage() error {
	msg, err := state.Client.Read()
	if err != nil {
		return err
	}

	if msg == nil {
		return nil
	}

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

// checkIntegrity compares sha1 hash sums of a piece and downloaded content
func (piece *TaskItem) checkIntegrity(content []byte) error {
	hash := sha1.Sum(content)
	if !bytes.Equal(hash[:], piece.Hash[:]) {
		return fmt.Errorf("piece %d failed integrity check", piece.Index)
	}

	return nil
}
