package torrent

import (
	"fmt"
	"log"

	"gihub.com/sauromates/leech/internal/utils"
	"github.com/schollz/progressbar/v3"
)

type Torrent struct {
	Peers       []utils.Peer
	PeerID      [20]byte
	InfoHash    [20]byte
	PieceHashes [][20]byte
	PieceLength int
	Length      int
	Name        string
}

type TaskItem struct {
	Index  int
	Hash   [20]byte
	Length int
}

type TaskResult struct {
	Index   int
	Content []byte
}

type TaskProgress struct {
	Index      int
	Client     *Client
	Content    []byte
	Downloaded int
	Requested  int
	Backlog    int
}

func (torrent *Torrent) Download(worker Worker) ([]byte, error) {
	log.Printf(
		"Starting to download \"%s\" from %d peers (%d pieces expected)\n",
		torrent.Name,
		len(torrent.Peers),
		len(torrent.PieceHashes),
	)

	taskQueue := make(chan *TaskItem, len(torrent.PieceHashes))
	resultQueue := make(chan *TaskResult)

	for index, hash := range torrent.PieceHashes {
		length := torrent.pieceSize(index)
		taskQueue <- &TaskItem{index, hash, length}
	}

	for _, peer := range torrent.Peers {
		go worker.Start(torrent, peer, taskQueue, resultQueue)
	}

	content := make([]byte, torrent.Length)
	bar := progressbar.Default(int64(len(torrent.PieceHashes)), "Downloading torrent")
	done := 0
	for done < len(torrent.PieceHashes) {
		res := <-resultQueue
		begin, end := torrent.pieceBounds(res.Index)

		copy(content[begin:end], res.Content)
		done++
		bar.Add(1)
	}

	close(taskQueue)
	bar.Finish()

	return content, nil
}

func (torrent *Torrent) pieceBounds(index int) (begin int, end int) {
	begin = index * torrent.PieceLength
	end = begin + torrent.PieceLength

	if end > torrent.Length {
		end = torrent.Length
	}

	return begin, end
}

func (torrent *Torrent) pieceSize(index int) int {
	begin, end := torrent.pieceBounds(index)

	return end - begin
}

func (state *TaskProgress) ReadMessage() error {
	msg, err := state.Client.Read()
	if err != nil {
		return fmt.Errorf("failed to read message (%s)", err)
	}

	if msg == nil {
		return nil
	}

	switch msg.ID {
	case MsgUnchoke:
		state.Client.IsChoked = false
	case MsgChoke:
		state.Client.IsChoked = true
	case MsgHave:
		index, err := ParseHave(msg)
		if err != nil {
			return err
		}

		state.Client.BitField.SetPiece(index)
	case MsgPiece:
		downloaded, err := ParsePiece(state.Index, state.Content, msg)
		if err != nil {
			return err
		}

		state.Downloaded += downloaded
		state.Backlog--
	}

	return nil
}
