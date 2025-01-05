package torrentfile

import (
	"os"

	"github.com/jackpal/bencode-go"
	"github.com/sauromates/leech/internal/utils"
)

type TorrentFile struct {
	Announce    string
	InfoHash    utils.BTString
	PieceHashes []utils.BTString
	PieceLength int
	Length      *int
	Name        string
	Paths       []utils.PathInfo
}

// Open unmarshals bencoded file into a TorrentFile struct
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

// GetLength returns actual torrent length for any concrete torrent type
// i.e. length of single-file torrent or sum of file sizes in multi-file torrent
func (tf *TorrentFile) GetLength() int {
	if len(tf.Paths) == 0 {
		return *tf.Length
	}

	// Last absolute length matches torrent length
	return tf.Paths[len(tf.Paths)-1].Length
}

// GetFiles creates a slice of files both for single- and multi-file torrents.
//
// In case of a single-file torrent it puts a file with name of the torrent
// itself with offset and length matched with the whole torrent [0:Length].
func (tf *TorrentFile) GetFiles() []utils.PathInfo {
	if len(tf.Paths) > 0 {
		return tf.Paths
	}

	return []utils.PathInfo{
		{Path: tf.Name, Offset: 0, Length: *tf.Length},
	}
}
