// Package worker is responsible for actual download process.
//
// It is designed to be used in non-blocking manner.
package worker

import (
	"errors"
	"log"

	"github.com/sauromates/leech/client"
	"github.com/sauromates/leech/internal/bthash"
	"github.com/sauromates/leech/internal/peers"
	"github.com/sauromates/leech/internal/piece"
)

// Returned upon rejected or timed out connection with a peer
var ErrConn error = errors.New("failed to connect to a peer")

// Worker holds info needed for connections with peers and downloading pieces
type Worker struct {
	peer     peers.Peer
	client   *client.Client
	infoHash bthash.Hash
	clientID bthash.Hash
}

// New creates new connection for a peer and puts it into new worker instance
func New(peer peers.Peer, infoHash, peerID bthash.Hash) *Worker {
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
func (w *Worker) Run(queue, results chan *piece.Piece) error {
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

		pipeline := NewPipeline(w.client, piece)
		if err := pipeline.Run(); err != nil {
			log.Printf("[ERROR] Download failed: %s", err)
			queue <- piece

			return err
		}

		if err := piece.VerifyHash(); err != nil {
			log.Printf("[ERROR] Invalid piece: %s", err)
			queue <- piece

			continue
		}

		w.client.ConfirmHavePiece(piece.Index)
		results <- piece
	}

	return nil
}
