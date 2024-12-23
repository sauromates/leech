package worker

import (
	"github.com/sauromates/leech/client"
	"github.com/sauromates/leech/internal/message"
	"github.com/sauromates/leech/internal/utils"
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

func (state *TaskProgress) Read() error {
	msg, err := state.Client.ReadMessage()
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
