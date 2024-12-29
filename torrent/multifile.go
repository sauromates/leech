package torrent

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/sauromates/leech/internal/peers"
	"github.com/sauromates/leech/internal/utils"
	"github.com/sauromates/leech/worker"
)

type MultiFileTorrent struct {
	BaseTorrent
	Paths []utils.FileInfo
}

// TotalSizeBytes sums length of all files in a multi file torrent.
func (torrent MultiFileTorrent) TotalSizeBytes() int {
	size := 0
	for _, file := range torrent.Paths {
		size += file.Length
	}

	return size
}

// Download runs workers asynchronously after preparing necessary infrastructure
// for them: assembles tasks and results queues, pushes peers into a pool of connections, etc.
func (torrent *MultiFileTorrent) Download(path string) ([]byte, error) {
	queue := make(chan *worker.Task, len(torrent.PieceHashes))
	results := make(chan *worker.TaskResult)
	pool := make(chan *peers.Peer, len(torrent.Peers))

	for index, hash := range torrent.PieceHashes {
		pieceLength := torrent.PieceSize(index)
		piece := worker.Task{Index: index, Hash: hash, Length: pieceLength}

		queue <- &piece
	}

	for _, peer := range torrent.Peers {
		pool <- &peer
	}

	done := 0
	for done < len(torrent.PieceHashes) {
		select {
		case piece := <-results:
			if err := torrent.Write(path, piece); err != nil {
				log.Fatal(err)
				return nil, err
			}

			done++
			percent := float64(done) / float64(len(torrent.PieceHashes)) * 100
			log.Printf("[INFO] Downloaded piece %d, %0.2f%% finished", piece.Index, percent)
		case peer := <-pool:
			log.Printf("[INFO] Received peer %s", peer.String())

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

	return []byte{}, nil
}

// Write populates a file with downloaded piece. BasePath is a directory where
// any downloading files should be stored.
func (torrent *MultiFileTorrent) Write(basePath string, piece *worker.TaskResult) error {
	filename, err := torrent.WhichFile(piece.Index)
	if err != nil {
		return err
	}

	filepath := filepath.Join(basePath, filename)
	file, err := os.OpenFile(filepath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, os.ModePerm)
	if err != nil {
		return err
	}

	defer file.Close()

	if _, err := file.Write(piece.Content); err != nil {
		return err
	}

	return nil
}

// WhichFile determines which file a piece belongs to based on its index
func (torrent *MultiFileTorrent) WhichFile(index int) (string, error) {
	type bound struct {
		name  string
		begin int
		end   int
	}

	// Assemble a slice of file boundaries
	bounds := make([]bound, len(torrent.Paths))
	for i, file := range torrent.Paths {
		var begin int
		if i == 0 {
			begin = 0
		} else {
			begin = bounds[i-1].end + 1
		}

		filename := filepath.Join(file.Path...)

		bounds[i] = bound{filename, begin, begin + file.Length}
	}

	pieceBegin, _ := torrent.PieceBounds(index)
	// Quickly search for the right interval for received piece index
	left, right := 0, len(bounds)-1
	for left <= right {
		mid := left + (right-left)/2
		file := bounds[mid]

		if pieceBegin >= file.begin && pieceBegin <= file.end {
			return file.name, nil
		} else if pieceBegin < file.begin {
			right = mid - 1
		} else {
			right = mid + 1
		}
	}

	return "", fmt.Errorf("failed to associate piece %d with a file", index)
}
