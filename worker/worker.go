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
		log.Printf("Could not handshake with %s (%s). Disconnecting\n", peer.IP, err)
		return nil, err
	}

	log.Printf("Connected to peer %s\n", peer.String())

	client.RequestUnchoke()
	client.AnnounceInterest()

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
			log.Printf("Download failed: %s", err)
			worker.queue <- piece

			return err
		}

		if err := piece.checkIntegrity(content); err != nil {
			log.Printf("Invalid piece: %s", err)
			worker.queue <- piece

			continue
		}

		worker.Peer.ConfirmHavePiece(piece.Index)
		worker.results <- &TaskResult{piece.Index, content}
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
			return nil, err
		}
	}

	return taskState.Content, nil
}
