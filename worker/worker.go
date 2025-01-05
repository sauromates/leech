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

// Worker holds info needed for connections with peers and downloading pieces
type Worker struct {
	peer     peers.Peer
	client   *client.Client
	infoHash utils.BTString
	clientID utils.BTString
}

// Create creates new connection for a peer and puts it into new worker instance
func Create(peer peers.Peer, infoHash, peerID utils.BTString) *Worker {
	return &Worker{peer, nil, infoHash, peerID}
}

// Connect opens new TCP connection with a peer
func (w *Worker) Connect() error {
	client, err := client.Create(w.peer, w.infoHash, w.clientID)
	if err != nil {
		return ErrConn
	}

	log.Printf("[INFO] Connected to peer %s\n", client.Conn.RemoteAddr())

	w.client = client

	return nil
}

// Run starts listening to a task queue until it's empty or until a download error occurs
func (w *Worker) Run(queue chan *Piece, results chan *PieceContent) error {
	if err := w.Connect(); err != nil {
		return err
	}

	defer w.client.Conn.Close()

	w.client.Unchoke()
	w.client.AnnounceInterest()

	for piece := range queue {
		if !w.client.BitField.HasPiece(piece.Index) {
			queue <- piece
			continue
		}

		content, err := w.downloadPiece(piece)
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

		w.client.ConfirmHavePiece(piece.Index)
		results <- &PieceContent{piece.Index, content}
	}

	return nil
}

// downloadPiece attempts to process given task by requesting pieces in a
// sequential pipeline
func (w *Worker) downloadPiece(piece *Piece) ([]byte, error) {
	pipeline := Pipeline{
		Index:   piece.Index,
		Client:  w.client,
		Content: make([]byte, piece.Length),
	}

	// Setting a deadline helps get unresponsive peers unstuck.
	// 30 seconds is more than enough time to download a 262 KB piece
	w.client.Conn.SetDeadline(time.Now().Add(30 * time.Second))
	defer w.client.Conn.SetDeadline(time.Time{})

	for pipeline.Downloaded < piece.Length {
		if !pipeline.Client.IsChoked {
			for pipeline.hasBacklogSpace(piece.Length) {
				blockSize := pipeline.blockSize(piece.Length)

				if err := w.client.RequestPiece(piece.Index, pipeline.Requested, blockSize); err != nil {
					return nil, err
				}

				pipeline.Backlog++
				pipeline.Requested += blockSize
			}
		}

		if err := pipeline.ReadMessage(); err != nil {
			return nil, err
		}
	}

	return pipeline.Content, nil
}
