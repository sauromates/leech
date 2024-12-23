package torrent

import (
	"log"
	"runtime"

	"github.com/sauromates/leech/internal/peers"
	"github.com/sauromates/leech/worker"
)

type SingleFileTorrent struct {
	BaseTorrent
}

func (torrent SingleFileTorrent) TotalSizeBytes() int {
	return torrent.Length
}

func (torrent *SingleFileTorrent) Download(path string) ([]byte, error) {
	queue := make(chan *worker.TaskItem, len(torrent.PieceHashes))
	results := make(chan *worker.TaskResult)
	pool := make(chan *peers.Peer, len(torrent.Peers))

	for i, hash := range torrent.PieceHashes {
		queue <- &worker.TaskItem{Index: i, Hash: hash, Length: torrent.PieceSize(i)}
	}

	for _, peer := range torrent.Peers {
		pool <- &peer
	}

	content, done := make([]byte, torrent.Length), 0
	for done < len(torrent.PieceHashes) {
		select {
		case piece := <-results:
			begin, end := torrent.PieceBounds(piece.Index)
			copy(content[begin:end], piece.Content)

			done++
		case peer := <-pool:
			numWorkers := runtime.NumGoroutine() - 1
			if numWorkers <= maxConnections {
				go torrent.startWorker(*peer, queue, results, pool)
			} else {
				pool <- peer
			}

			log.Printf("Started worker for peer %s, currently %d are running", peer.String(), numWorkers)
		}
	}

	close(queue)
	close(results)
	close(pool)

	return content, nil
}
