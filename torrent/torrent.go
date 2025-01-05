package torrent

import (
	"github.com/sauromates/leech/internal/peers"
	"github.com/sauromates/leech/internal/utils"
	"github.com/sauromates/leech/worker"
)

// maxConnections defines a limit of peer connections
const maxConnections int = 10

type Torrent interface {
	// TotalSizeBytes calculates total size of downloadable content for torrent
	TotalSizeBytes() int
	// Download performs all necessary operations to download torrent contents
	Download(path string) ([]byte, error)
}

type BaseTorrent struct {
	Peers       []peers.Peer
	PeerID      utils.BTString
	InfoHash    utils.BTString
	PieceHashes []utils.BTString
	PieceLength int
	Name        string
	Length      int
}

// PieceBounds calculates where the piece with given index begins
// and ends within the torrent contents
func (torrent *BaseTorrent) PieceBounds(index int) (begin int, end int) {
	begin = index * torrent.PieceLength
	end = begin + torrent.PieceLength

	if end > torrent.Length {
		end = torrent.Length
	}

	return begin, end
}

// PieceSize returns the size of a piece with given index in bytes
func (torrent *BaseTorrent) PieceSize(index int) int {
	begin, end := torrent.PieceBounds(index)

	return end - begin
}

// startWorker transforms peer into a listener for a task queue and
// runs it until the queue is empty or until an error occurs.
func (torrent *BaseTorrent) startWorker(
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
