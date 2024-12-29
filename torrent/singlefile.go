package torrent

import (
	"log"
	"runtime"
	"time"

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
	queue := make(chan *worker.Task, len(torrent.PieceHashes))
	results := make(chan *worker.TaskResult)
	pool := make(chan *peers.Peer, len(torrent.Peers))

	for i, hash := range torrent.PieceHashes {
		queue <- &worker.Task{Index: i, Hash: hash, Length: torrent.PieceSize(i)}
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
			percent := float64(done) / float64(len(torrent.PieceHashes)) * 100
			log.Printf("[INFO] Downloaded piece %d, %0.2f%% finished", piece.Index, percent)
		case peer := <-pool:
			numWorkers := runtime.NumGoroutine() - 1
			if numWorkers < maxConnections {
				log.Printf("[INFO] Connecting to %s", peer.String())
				go torrent.startWorker(*peer, queue, results, pool)
			} else {
				log.Printf("[INFO] Too many connections, will retry %s later", peer.String())
				time.Sleep(time.Second * 3)
				pool <- peer
			}
		}
	}

	close(queue)
	close(results)
	close(pool)

	return content, nil
}
