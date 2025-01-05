package torrentfile

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"io"
	"path/filepath"

	"github.com/jackpal/bencode-go"
	"github.com/sauromates/leech/internal/utils"
)

type bencodeFile struct {
	Length int      `bencode:"length"`
	Path   []string `bencode:"path"`
}

type bencodeInfo struct {
	Pieces      string        `bencode:"pieces"`
	PieceLength int           `bencode:"piece length"`
	Length      int           `bencode:"length"`
	Name        string        `bencode:"name"`
	Files       []bencodeFile `bencode:"files"`
}

type singleFileBencodeInfo struct {
	Pieces      string `bencode:"pieces"`
	PieceLength int    `bencode:"piece length"`
	Length      int    `bencode:"length"`
	Name        string `bencode:"name"`
}

type multiFileBencodeInfo struct {
	Pieces      string        `bencode:"pieces"`
	PieceLength int           `bencode:"piece length"`
	Name        string        `bencode:"name"`
	Files       []bencodeFile `bencode:"files"`
}

type bencodeTorrent struct {
	Announce string      `bencode:"announce"`
	Comment  string      `bencode:"comment"`
	Info     bencodeInfo `bencode:"info"`
}

// Decodes torrent file contents via `bencode` module
func DecodeTorrentFile(reader io.Reader) (*bencodeTorrent, error) {
	torrent := bencodeTorrent{}
	if err := bencode.Unmarshal(reader, &torrent); err != nil {
		return nil, err
	}

	return &torrent, nil
}

// Hashes whole torrent info via sha1.
func (info *bencodeInfo) hash() (utils.BTString, error) {
	var buffer bytes.Buffer
	var hashableInfo interface{}

	if len(info.Files) == 0 {
		hashableInfo = singleFileBencodeInfo{
			Pieces:      info.Pieces,
			PieceLength: info.PieceLength,
			Length:      info.Length,
			Name:        info.Name,
		}
	} else {
		hashableInfo = multiFileBencodeInfo{
			Pieces:      info.Pieces,
			PieceLength: info.PieceLength,
			Name:        info.Name,
			Files:       info.Files,
		}
	}

	if err := bencode.Marshal(&buffer, hashableInfo); err != nil {
		return utils.BTString{}, err
	}

	return sha1.Sum(buffer.Bytes()), nil
}

// Creates a hash for each parsed piece and wraps them all in a slice
// resulting in infohash used to uniquely identify file
func (info *bencodeInfo) hashPieces() ([]utils.BTString, error) {
	buffer, hashLen := []byte(info.Pieces), 20

	if len(buffer)%hashLen != 0 {
		return nil, fmt.Errorf("received malformed pieces of length %d", len(buffer))
	}

	// Calculate how many pieces there are by splitting the whole string by 20 bytes
	hashCount := len(buffer) / hashLen
	hashes := make([]utils.BTString, hashCount)

	// Iterate over each 20 byte chunk and put it into the slice of hashes
	for i := 0; i < hashCount; i++ {
		begin, end := i*hashLen, (i+1)*hashLen
		copy(hashes[i][:], buffer[begin:end])
	}

	return hashes, nil
}

// fileBounds calculates the offset and length of each file in the torrent
func (info *bencodeInfo) fileBounds() []utils.PathInfo {
	bounds := make([]utils.PathInfo, len(info.Files))
	offset := 0

	for i, file := range info.Files {
		bounds[i] = utils.PathInfo{
			Path:   filepath.Join(file.Path...),
			Offset: offset,
			Length: offset + file.Length,
		}

		offset += file.Length
	}

	return bounds
}

func (torrent *bencodeTorrent) createTorrentFile() (TorrentFile, error) {
	infoHash, err := torrent.Info.hash()
	if err != nil {
		return TorrentFile{}, err
	}

	pieceHashes, err := torrent.Info.hashPieces()
	if err != nil {
		return TorrentFile{}, err
	}

	length := 0
	for _, file := range torrent.Info.Files {
		length += file.Length
	}

	file := TorrentFile{
		Announce:    torrent.Announce,
		InfoHash:    infoHash,
		PieceHashes: pieceHashes,
		PieceLength: torrent.Info.PieceLength,
		Length:      &length,
		Name:        torrent.Info.Name,
		Paths:       torrent.Info.fileBounds(),
	}

	return file, nil
}
