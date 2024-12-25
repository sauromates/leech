package worker

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"log"
	"time"

	"github.com/sauromates/leech/client"
)

const (
	// MaxBlockSize is the largest number of bytes a request can ask for (default is 2048 Kb (16384 bytes))
	MaxBlockSize = 16384
	// MaxBacklog is the number of unfulfilled requests a client can have in its pipeline
	MaxBacklog = 50
)

type Worker struct {
	Peer *client.Client
}

// Run starts listening to a task queue until it's empty or until a
// download error occurs
func (worker *Worker) Run(queue chan *TaskItem, results chan *TaskResult) error {
	for piece := range queue {
		if !worker.Peer.BitField.HasPiece(piece.Index) {
			queue <- piece
			continue
		}

		content, err := worker.downloadPiece(piece)
		if err != nil {
			log.Printf("Download failed: %s", err)
			queue <- piece

			return err
		}

		if err := verifyPiece(piece, content); err != nil {
			log.Printf("Integrity check failed: %s", err)
			queue <- piece

			continue
		}

		worker.Peer.ConfirmHavePiece(piece.Index)

		results <- &TaskResult{piece.Index, content}
	}

	return nil
}

func (worker *Worker) downloadPiece(piece *TaskItem) ([]byte, error) {
	taskState := TaskProgress{
		Index:   piece.Index,
		Client:  worker.Peer,
		Content: make([]byte, piece.Length),
	}

	// Setting a deadline helps get unresponsive peers unstuck.
	// 30 seconds is more than enough time to download a 262 KB piece
	worker.Peer.Conn.SetDeadline(time.Now().Add(30 * time.Second))
	defer worker.Peer.Conn.SetDeadline(time.Time{})

	for taskState.Downloaded < piece.Length {
		if !taskState.Client.IsChoked {
			for taskState.Backlog < MaxBacklog && taskState.Requested < piece.Length {
				blockSize := MaxBlockSize
				currBlockSize := piece.Length - taskState.Requested
				if currBlockSize < blockSize {
					blockSize = currBlockSize
				}

				if err := worker.Peer.RequestPiece(piece.Index, taskState.Requested, blockSize); err != nil {
					return nil, err
				}

				taskState.Backlog++
				taskState.Requested += blockSize
			}
		}

		if err := taskState.ReadMessage(); err != nil {
			log.Fatal("here")
			return nil, err
		}
	}

	return taskState.Content, nil
}
