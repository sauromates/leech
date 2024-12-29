package torrent

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/sauromates/leech/internal/peers"
	"github.com/sauromates/leech/internal/utils"
	"github.com/sauromates/leech/worker"
	"github.com/schollz/progressbar/v3"
)

type FileWithBounds struct {
	name  string
	begin int
	end   int
}

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

	done, bar := 0, progressbar.DefaultBytes(int64(torrent.TotalSizeBytes()), "downloading")
	for done < len(torrent.PieceHashes) {
		select {
		case piece := <-results:
			if err := torrent.Write(path, piece, bar); err != nil {
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

	bar.Finish()

	return []byte{}, nil
}

// Write populates a file with downloaded piece. BasePath is a directory where
// any downloaded files should be stored.
func (torrent *MultiFileTorrent) Write(basePath string, piece *worker.TaskResult, tracker io.Writer) error {
	fileBounds, err := torrent.WhichFile(piece.Index)
	if err != nil {
		return err
	}

	filepath := filepath.Join(basePath, fileBounds.name)
	file, err := os.OpenFile(filepath, os.O_CREATE|os.O_WRONLY, os.ModePerm)
	if err != nil {
		return err
	}

	defer file.Close()

	writer := io.NewOffsetWriter(file, int64(fileBounds.begin))
	if _, err := io.Copy(io.MultiWriter(writer, tracker), bytes.NewReader(piece.Content)); err != nil {
		return err
	}

	return nil
}

// WhichFile determines which file a piece belongs to based on its index
func (torrent *MultiFileTorrent) WhichFile(index int) (FileWithBounds, error) {
	// Assemble a slice of file boundaries
	bounds := make([]FileWithBounds, len(torrent.Paths))
	cumulativeOffset := 0
	for i, file := range torrent.Paths {
		bounds[i] = FileWithBounds{
			name:  filepath.Join(file.Path...),
			begin: cumulativeOffset,
			end:   cumulativeOffset + file.Length + 1,
		}

		cumulativeOffset += file.Length
	}

	pieceBegin, pieceEnd := torrent.PieceBounds(index)
	// Quickly search for the right interval for received piece index
	left, right := 0, len(bounds)-1
	for left <= right {
		mid := left + (right-left)/2
		file := bounds[mid]

		if pieceBegin <= file.end && pieceEnd >= file.begin {
			return FileWithBounds{
				name:  file.name,
				begin: max(0, pieceBegin-file.begin),
				end:   min(file.end-file.begin+1, pieceEnd-file.begin),
			}, nil
		} else if pieceBegin < file.begin {
			right = mid - 1
		} else {
			left = mid + 1
		}
	}

	return FileWithBounds{}, fmt.Errorf("failed to associate piece %d with a file", index)
}
