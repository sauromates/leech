package torrentfile

import (
	"crypto/rand"
	"os"
	"path/filepath"

	"github.com/jackpal/bencode-go"
	"github.com/sauromates/leech/internal/utils"
	"github.com/sauromates/leech/torrent"
)

type TorrentFile struct {
	Announce    string
	InfoHash    utils.BTString
	PieceHashes []utils.BTString
	PieceLength int
	Length      *int
	Name        string
	Paths       []utils.FileInfo
}

func Open(path string) (TorrentFile, error) {
	file, err := os.Open(path)
	if err != nil {
		return TorrentFile{}, err
	}

	defer file.Close()

	torrent := bencodeTorrent{}
	if err := bencode.Unmarshal(file, &torrent); err != nil {
		return TorrentFile{}, err
	}

	return torrent.createTorrentFile()
}

func (tf *TorrentFile) Parse() (torrent.Torrent, error) {
	var peerID utils.BTString
	if _, err := rand.Read(peerID[:]); err != nil {
		return nil, err
	}

	peers, err := tf.requestPeers(peerID, uint16(49160))
	if err != nil {
		return nil, err
	}

	baseMeta := torrent.BaseTorrent{
		Peers:       peers,
		PeerID:      peerID,
		InfoHash:    tf.InfoHash,
		PieceHashes: tf.PieceHashes,
		PieceLength: tf.PieceLength,
		Name:        tf.Name,
		Length:      tf.GetLength(),
	}

	var parsed torrent.Torrent
	if len(tf.Paths) > 0 {
		parsed = &torrent.MultiFileTorrent{BaseTorrent: baseMeta, Paths: tf.Paths}
	} else {
		parsed = &torrent.SingleFileTorrent{BaseTorrent: baseMeta}
	}

	return parsed, nil
}

func (tfile *TorrentFile) Download(to string) error {
	source, err := tfile.Parse()
	if err != nil {
		return err
	}

	content, err := source.Download(filepath.Dir(to))
	if err != nil {
		return err
	}

	outFile, err := os.Create(to)
	if err != nil {
		return err
	}

	defer outFile.Close()

	if _, err := outFile.Write(content); err != nil {
		return err
	}

	return nil
}

func (tfile *TorrentFile) DownloadMultiple(dir string) error {
	source, err := tfile.Parse()
	if err != nil {
		return err
	}

	if _, err := source.Download(dir); err != nil {
		return err
	}

	return nil
}

func (torrent *TorrentFile) GetLength() int {
	if len(torrent.Paths) == 0 {
		return *torrent.Length
	}

	size := 0
	for _, file := range torrent.Paths {
		size += file.Length
	}

	return size
}
