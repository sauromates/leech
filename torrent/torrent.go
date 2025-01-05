package torrent

import (
	"fmt"
	"io"
	"log"
	"path/filepath"
	"runtime"
	"time"

	"github.com/sauromates/leech/internal/peers"
	"github.com/sauromates/leech/internal/utils"
	"github.com/sauromates/leech/torrentfile"
	"github.com/sauromates/leech/worker"
	"github.com/schollz/progressbar/v3"
)

const (
	appName        string = "leech"
	maxConnections int    = 10
)

type Torrent struct {
	Peers       []peers.Peer
	PeerID      utils.BTString
	InfoHash    utils.BTString
	PieceHashes []utils.BTString
	PieceLength int
	Name        string
	Length      int
	Files       []utils.PathInfo
	DownloadDir string
}

func CreateFromTorrentFile(tf torrentfile.TorrentFile) (*Torrent, error) {
	var peerID utils.BTString
	copy(peerID[:], appName)

	peers, err := tf.RequestPeers(peerID, uint16(49160))
	if err != nil {
		return nil, err
	}

	torrent := Torrent{
		Peers:       peers,
		PeerID:      peerID,
		InfoHash:    tf.InfoHash,
		PieceHashes: tf.PieceHashes,
		PieceLength: tf.PieceLength,
		Name:        tf.Name,
		Length:      tf.GetLength(),
		Files:       tf.GetFiles(),
	}

	return &torrent, nil
}

// Download runs workers asynchronously after preparing necessary infrastructure
// for them: assembles tasks and results queues, pushes peers into a pool of connections, etc.
func (torrent *Torrent) Download(dir string) error {
	torrent.DownloadDir = dir

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
	tracker := progressbar.DefaultBytes(int64(torrent.Length), "Downloading")
	for len(done) < len(torrent.PieceHashes) {
		select {
		case piece := <-results:
			// Skip if a piece was marked as done
			if done[piece.Index] {
				continue
			}

			if _, err := torrent.Write(piece, tracker); err != nil {
				return err
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

	return tracker.Finish()
}

// Write copies received piece into associated files.
//
// Base scenario is writing to a single file, but pieces may overlap files,
// in which case we will split piece by relative offset and length and write
// its parts to multiple associated files
func (torrent *Torrent) Write(piece *worker.TaskResult, tracker io.Writer) (n int, err error) {
	files, err := torrent.WhichFiles(piece.Index)
	if err != nil {
		return n, err
	}

	for _, file := range files {
		file.Open()

		src := io.NewSectionReader(piece, file.PieceStart, file.PieceEnd)
		var dst io.Writer = file
		if tracker != nil {
			dst = io.MultiWriter(dst, tracker)
		}

		written, err := io.Copy(dst, src)
		if err != nil {
			file.Close()
			return int(n), err
		}

		file.Close()

		n += int(written)
	}

	return n, err
}

// WhichFiles determines which files the piece belongs to by an intersection
// of absolute offsets and lengths.
func (torrent *Torrent) WhichFiles(piece int) (map[string]utils.FileMap, error) {
	files := make(map[string]utils.FileMap)
	offset, length := torrent.PieceBounds(piece)

	for _, file := range torrent.Files {
		intersectOffset := max(offset, file.Offset)
		intersectLength := min(length, file.Length)

		if intersectOffset < intersectLength {
			relativeOffset := intersectOffset - offset
			relativeLength := intersectLength - intersectOffset

			files[file.Path] = utils.FileMap{
				FileName:   filepath.Join(torrent.DownloadDir, file.Path),
				FileOffset: int64(intersectOffset - file.Offset),
				PieceStart: int64(relativeOffset),
				PieceEnd:   int64(relativeOffset + relativeLength),
			}
		}
	}

	if len(files) == 0 {
		return nil, fmt.Errorf("no files found for piece %d", piece)
	}

	return files, nil
}

// PieceBounds calculates where the piece with given index begins
// and ends within the torrent contents
func (torrent *Torrent) PieceBounds(index int) (begin int, end int) {
	begin = index * torrent.PieceLength
	end = begin + torrent.PieceLength

	if end > torrent.Length {
		end = torrent.Length
	}

	return begin, end
}

// PieceSize returns the size of a piece with given index in bytes
func (torrent *Torrent) PieceSize(index int) int {
	begin, end := torrent.PieceBounds(index)

	return end - begin
}

func (t Torrent) String() string {
	return fmt.Sprintf("Torrent %s\n---\nTotalSize: %.2f MB\nTotalFiles: %d\n",
		t.Name,
		float64(t.Length)/(1024*1024),
		len(t.Files),
	)
}

// startWorker transforms peer into a listener for a task queue and
// runs it until the queue is empty or until an error occurs.
func (torrent *Torrent) startWorker(
	peer peers.Peer,
	queue chan *worker.Task,
	results chan *worker.TaskResult,
	peers chan *peers.Peer,
) {
	w := worker.Create(peer, torrent.InfoHash, torrent.PeerID)

	if err := w.Run(queue, results); err != nil {
		// Peer should be put back to the pool only if there was no
		// connection errors, otherwise it's pointless since workers would
		// constantly try to reconnect to unresponsive peers
		if err != worker.ErrConn {
			peers <- &peer
		}
	}
}
