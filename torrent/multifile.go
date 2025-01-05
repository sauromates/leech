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

type MultiFileTorrent struct {
	BaseTorrent
	Paths []utils.PathInfo
}

type FileMap struct {
	FileOffset int64
	PieceStart int
	PieceEnd   int
}

// TotalSizeBytes sums length of all files in a multi file torrent.
func (torrent MultiFileTorrent) TotalSizeBytes() int {
	return torrent.Length
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

	done := make(map[int]bool)
	tracker := progressbar.DefaultBytes(int64(torrent.TotalSizeBytes()), "Downloading")
	for len(done) < len(torrent.PieceHashes) {
		select {
		case piece := <-results:
			// Skip if a piece was marked as done
			if done[piece.Index] {
				continue
			}

			if _, err := torrent.Write(path, piece, tracker); err != nil {
				return nil, err
			}

			done[piece.Index] = true
			percent := float64(len(done)) / float64(len(torrent.PieceHashes)) * 100
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

	tracker.Finish()

	return []byte{}, nil
}

// Write copies received piece into associated files.
//
// Base scenario is writing to a single file, but pieces may overlap files,
// in which case we will split piece by relative offset and length and write
// its parts to multiple associated files
func (torrent *MultiFileTorrent) Write(basePath string, piece *worker.TaskResult, tracker io.Writer) (int, error) {
	files, err := torrent.WhichFiles(piece.Index)
	if err != nil {
		return 0, err
	}

	total := 0

	for path, chunk := range files {
		filepath := filepath.Join(basePath, path)
		file, err := os.OpenFile(filepath, os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return 0, err
		}

		defer file.Close()

		length := chunk.PieceEnd - chunk.PieceStart
		content := make([]byte, length)
		copy(content, piece.Content[chunk.PieceStart:chunk.PieceEnd])

		src := bytes.NewReader(content)
		var dst io.Writer = io.NewOffsetWriter(file, chunk.FileOffset)
		if tracker != nil {
			dst = io.MultiWriter(dst, tracker)
		}

		copied, err := io.Copy(dst, src)
		if err != nil {
			return 0, err
		}

		file.Close()

		total += int(copied)
	}

	return total, nil
}

// WhichFiles determines which files the piece belongs to by an intersection
// of absolute offsets and lengths.
func (torrent *MultiFileTorrent) WhichFiles(piece int) (map[string]FileMap, error) {
	files := make(map[string]FileMap)
	offset, length := torrent.PieceBounds(piece)

	for _, file := range torrent.Paths {
		intersectOffset := max(offset, file.Offset)
		intersectLength := min(length, file.Length)

		if intersectOffset < intersectLength {
			relativeOffset := intersectOffset - offset
			relativeLength := intersectLength - intersectOffset

			files[file.Path] = FileMap{
				FileOffset: int64(intersectOffset - file.Offset),
				PieceStart: relativeOffset,
				PieceEnd:   relativeOffset + relativeLength,
			}
		}
	}

	if len(files) == 0 {
		return nil, fmt.Errorf("no files found for piece %d", piece)
	}

	return files, nil
}
