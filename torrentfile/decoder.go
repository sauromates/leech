package torrentfile

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"io"

	"gihub.com/sauromates/leech/internal/utils"
	"github.com/jackpal/bencode-go"
)

type bencodeInfo struct {
	Pieces      string           `bencode:"pieces"`
	PieceLength int              `bencode:"piece length"`
	Length      int              `bencode:"length"`
	Name        string           `bencode:"name"`
	Files       []utils.FileInfo `bencode:"files"`
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
func (info *bencodeInfo) hash() ([20]byte, error) {
	var buffer bytes.Buffer
	if err := bencode.Marshal(&buffer, *info); err != nil {
		return [20]byte{}, err
	}

	return sha1.Sum(buffer.Bytes()), nil
}

// Creates a hash for each parsed piece and wraps them all in a slice
// resulting in infohash used to uniquely identify file
func (info *bencodeInfo) hashPieces() ([][20]byte, error) {
	buffer, hashLen := []byte(info.Pieces), 20

	if len(buffer)%hashLen != 0 {
		return nil, fmt.Errorf("Received malformed pieces of length %d", len(buffer))
	}

	hashCount := len(buffer) / hashLen
	hashes := make([][20]byte, hashCount)

	for i := 0; i < hashCount; i++ {
		// I have no idea what is going on here
		// @todo return later
		copy(hashes[i][:], buffer[i*hashLen:(i+1)*hashLen])
	}

	return hashes, nil
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

	file := TorrentFile{
		Announce:    torrent.Announce,
		InfoHash:    infoHash,
		PieceHashes: pieceHashes,
		PieceLength: torrent.Info.PieceLength,
		Length:      torrent.Info.Length,
		Name:        torrent.Info.Name,
		Paths:       torrent.Info.Files,
	}

	return file, nil
}
