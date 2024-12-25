package worker

import (
	"log"
	"time"

	"github.com/sauromates/leech/client"
	"github.com/sauromates/leech/internal/peers"
	"github.com/sauromates/leech/internal/utils"
)

// Worker listens to tasks queue and processes each received task.
// Peer field here is actually a connection here.
type Worker struct {
	Peer    *client.Client
	queue   chan *TaskItem
	results chan *TaskResult
}

// TorrentConnInfo holds information about the torrent unique hash ID for
// identification and current user ID for connections
type TorrentConnInfo struct {
	InfoHash utils.BTString
	MyID     utils.BTString
}

// Create creates new connection for a peer and puts it into new worker instance
func Create(ti TorrentConnInfo, peer peers.Peer, queue chan *TaskItem, results chan *TaskResult) (*Worker, error) {
	client, err := client.Create(peer, ti.InfoHash, ti.MyID)
	if err != nil {
		return nil, err
	}

	log.Printf("[INFO] Connected to peer %s\n", peer.String())

	return &Worker{client, queue, results}, nil
}

// Run starts listening to a task queue until it's empty or until a download error occurs
func (worker *Worker) Run() error {
	defer worker.Peer.Conn.Close()

	worker.Peer.Unchoke()
	worker.Peer.AnnounceInterest()

	for piece := range worker.queue {
		if !worker.Peer.BitField.HasPiece(piece.Index) {
			worker.queue <- piece
			continue
		}

		content, err := worker.downloadPiece(piece)
		if err != nil {
			log.Printf("[ERROR] Download failed: %s", err)
			worker.queue <- piece

			return err
		}

		if err := piece.checkIntegrity(content); err != nil {
			log.Printf("[ERROR] Invalid piece: %s", err)
			worker.queue <- piece

			continue
		}

		worker.Peer.ConfirmHavePiece(piece.Index)
		worker.results <- &TaskResult{piece.Index, content}
	}

	return nil
}

func (worker *Worker) downloadPiece(piece *TaskItem) ([]byte, error) {
	task := TaskProgress{
		Index:   piece.Index,
		Client:  worker.Peer,
		Content: make([]byte, piece.Length),
	}

	// Setting a deadline helps get unresponsive peers unstuck.
	// 30 seconds is more than enough time to download a 262 KB piece
	worker.Peer.Conn.SetDeadline(time.Now().Add(30 * time.Second))
	defer worker.Peer.Conn.SetDeadline(time.Time{})

	for task.Downloaded < piece.Length {
		if !worker.Peer.IsChoked {
			for task.hasBacklogSpace(piece) {
				blockSize := task.blockSize(piece)

				if err := worker.Peer.RequestPiece(piece.Index, task.Requested, blockSize); err != nil {
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
