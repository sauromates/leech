package torrent

import (
	"log"

	"github.com/sauromates/leech/client"
	"github.com/sauromates/leech/internal/peers"
	"github.com/sauromates/leech/internal/utils"
	"github.com/sauromates/leech/worker"
)

const maxConnections int = 5

type Torrent interface {
	// TotalSizeBytes calculates total size of downloadable content for torrent
	TotalSizeBytes() int
	Download(path string) ([]byte, error)
	ReadInfoHash() utils.BTString
	ReadPeerID() utils.BTString
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

func (torrent *BaseTorrent) ReadInfoHash() utils.BTString {
	return torrent.InfoHash
}

func (torrent *BaseTorrent) ReadPeerID() utils.BTString {
	return torrent.PeerID
}

func (torrent *BaseTorrent) PieceBounds(index int) (begin int, end int) {
	begin = index * torrent.PieceLength
	end = begin + torrent.PieceLength

	if end > torrent.Length {
		end = torrent.Length
	}

	return begin, end
}

func (torrent *BaseTorrent) PieceSize(index int) int {
	begin, end := torrent.PieceBounds(index)

	return end - begin
}

func (torrent *BaseTorrent) startWorker(peer peers.Peer, queue chan *worker.TaskItem, results chan *worker.TaskResult, peers chan *peers.Peer) {
	client, err := client.Create(peer, torrent.ReadInfoHash(), torrent.ReadPeerID())
	if err != nil {
		log.Printf("Could not handshake with %s (%s). Disconnecting\n", peer.IP, err)
		return
	}

	log.Printf("Completed handshake with %s\n", peer.IP)

	defer client.Conn.Close()

	client.RequestUnchoke()
	client.AnnounceInterest()

	worker := &worker.Worker{Peer: client}
	if err := worker.Run(queue, results); err != nil {
		peers <- &peer
	}
}
