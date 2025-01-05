package worker

import (
	"errors"
	"log"
	"time"

	"github.com/sauromates/leech/client"
	"github.com/sauromates/leech/internal/peers"
	"github.com/sauromates/leech/internal/utils"
)

// Returned upon rejected or timed out connection with a peer
var ErrConn error = errors.New("failed to connect to a peer")

// Worker listens to tasks queue and processes each received task.
// Peer field is actually a connection here.
type Worker struct {
	InfoHash utils.BTString
	ClientID utils.BTString
}

// Create creates new connection for a peer and puts it into new worker instance
func Create(infoHash, peerID utils.BTString) *Worker {
	return &Worker{infoHash, peerID}
}

// Connect opens new TCP connection with given peer
func (worker *Worker) Connect(peer peers.Peer) (*client.Client, error) {
	client, err := client.Create(peer, worker.InfoHash, worker.ClientID)
	if err != nil {
		return nil, ErrConn
	}

	log.Printf("[INFO] Connected to peer %s\n", client.Conn.RemoteAddr())

	return client, nil
}

// Run starts listening to a task queue until it's empty or until a download error occurs
func (worker *Worker) Run(peer peers.Peer, queue chan *Task, results chan *TaskResult) error {
	client, err := worker.Connect(peer)
	if err != nil {
		return err
	}

	defer client.Conn.Close()

	client.Unchoke()
	client.AnnounceInterest()

	for piece := range queue {
		if !client.BitField.HasPiece(piece.Index) {
			queue <- piece
			continue
		}

		content, err := worker.downloadPiece(client, piece)
		if err != nil {
			log.Printf("[ERROR] Download failed: %s", err)
			queue <- piece

			return err
		}

		if err := piece.verifyHashSum(content); err != nil {
			log.Printf("[ERROR] Invalid piece: %s", err)
			queue <- piece

			continue
		}

		client.ConfirmHavePiece(piece.Index)
		results <- &TaskResult{piece.Index, content}
	}

	return nil
}

// downloadPiece attempts to process given task by requesting pieces in a
// sequential pipeline
func (worker *Worker) downloadPiece(client *client.Client, piece *Task) ([]byte, error) {
	task := TaskProgress{
		Index:   piece.Index,
		Client:  client,
		Content: make([]byte, piece.Length),
	}

	// Setting a deadline helps get unresponsive peers unstuck.
	// 30 seconds is more than enough time to download a 262 KB piece
	client.Conn.SetDeadline(time.Now().Add(30 * time.Second))
	defer client.Conn.SetDeadline(time.Time{})

	for task.Downloaded < piece.Length {
		if !task.Client.IsChoked {
			for task.hasBacklogSpace(piece.Length) {
				blockSize := task.blockSize(piece.Length)

				if err := client.RequestPiece(piece.Index, task.Requested, blockSize); err != nil {
					return nil, err
				}

				task.Backlog++
				task.Requested += blockSize
			}
		}

		if err := task.ReadMessage(); err != nil {
			return nil, err
		}
	}

	return task.Content, nil
}
