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

func (torrent MultiFileTorrent) TotalSizeBytes() int {
	size := 0
	for _, file := range torrent.Paths {
		size += file.Length
	}

	return size
}

func (torrent *MultiFileTorrent) Download(path string) ([]byte, error) {
	queue := make(chan *worker.TaskItem, len(torrent.PieceHashes))
	results := make(chan *worker.TaskResult)
	pool := make(chan *peers.Peer, len(torrent.Peers))

	for i, hash := range torrent.PieceHashes {
		queue <- &worker.TaskItem{Index: i, Hash: hash, Length: torrent.PieceSize(i)}
	}

	for _, peer := range torrent.Peers {
		pool <- &peer
	}

	done := 0
	for done < len(torrent.PieceHashes) {
		select {
		case piece := <-results:
			if err := torrent.Write(path, piece); err != nil {
				return nil, err
			}

			done++
		case peer := <-pool:
			numWorkers := runtime.NumGoroutine() - 1
			if numWorkers <= maxConnections {
				go torrent.startWorker(*peer, queue, results, pool)
			} else {
				time.Sleep(5 * time.Second)
				pool <- peer
			}

			log.Printf("Started worker for peer %s, currently %d are running", peer.String(), numWorkers)
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
