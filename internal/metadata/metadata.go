// Package metadata provides support for creating and manipulating torrent
// metadata from different sources like `.torrent` files and magnet links.
//
// Metadata is not meant to be used outside of its package except for initial
// parsing hence unexported structs and methods.
package metadata

import (
	"bytes"
	"crypto/sha1"
	"errors"
	"fmt"
	"io"
	"path/filepath"

	"github.com/jackpal/bencode-go"
	"github.com/sauromates/leech/internal/bthash"
	"github.com/sauromates/leech/internal/peers"
	"github.com/sauromates/leech/internal/piece"
	"github.com/sauromates/leech/internal/utils"
	"github.com/sauromates/leech/torrent"
	"github.com/sauromates/leech/tracker"
)

// file represents single entry in torrent's `files` dictionary.
type file struct {
	Length int      `bencode:"length"`
	Path   []string `bencode:"path"`
}

// info represents common torrent's metadata and is a crucial part of
// our knowledge about the torrent.
type info struct {
	Pieces      string `bencode:"pieces"`
	PieceLength int    `bencode:"piece length"`
	Length      int    `bencode:"length"`
	Name        string `bencode:"name"`
	Files       []file `bencode:"files"`
}

// Metadata is a top-level dictionary holding torrent info. [Metadata.Announce]
// and [Metadata.Comment] fields are generally parsed from `.torrent` files
// and are missing from metadata exchange data.
type Metadata struct {
	Announce string `bencode:"announce"`
	Comment  string `bencode:"comment"`
	Info     info   `bencode:"info"`
}

// Parse attempts to unmarshal given source via [bencode.Unmarshal] and returns
// parsed struct on success or unmarshaling error.
func Parse(source io.Reader) (*Metadata, error) {
	md := Metadata{}
	if err := bencode.Unmarshal(source, &md); err != nil {
		return nil, err
	}

	return &md, nil
}

// NewTorrent transforms metadata into [torrent.Torrent] struct.
func (m *Metadata) NewTorrent(client *peers.Peer) (*torrent.Torrent, error) {
	if m == nil {
		return nil, errors.New("empty metadata")
	}

	pieces, err := m.Info.hashPieces()
	if err != nil {
		return nil, err
	}

	infoHash, err := m.Info.hash()
	if err != nil {
		return nil, err
	}

	t := torrent.Torrent{
		Name:     m.Info.Name,
		Length:   m.Info.getLength(),
		InfoHash: infoHash,
		Client:   client,
		Pieces:   pieces,
		Files:    m.Info.getFiles(),
	}

	t.Tracker = tracker.New(m.Announce, infoHash, client, t.Length)

	return &t, nil
}

// hash generates SHA-1 hash sum of bencoded torrent metadata - "info hash".
//
// Since torrent metadata may hold either `length` or `files` field but never
// both it's important to remove absent field from a struct first because
// even nil value will affect the resulted hash sum making it invalid.
func (i *info) hash() (bthash.Hash, error) {
	metadata := map[string]any{
		"pieces":       i.Pieces,
		"piece length": i.PieceLength,
		"length":       i.Length,
		"name":         i.Name,
		"files":        i.Files,
	}

	if len(i.Files) == 0 {
		delete(metadata, "files")
	} else {
		delete(metadata, "length")
	}

	var buf bytes.Buffer
	if err := bencode.Marshal(&buf, metadata); err != nil {
		return bthash.Hash{}, err
	}

	return sha1.Sum(buf.Bytes()), nil
}

// hashPieces splits raw `pieces` string from torrent's metadata into a slice
// of [20]byte hashes and transforms each into [piece.Piece] struct.
func (i *info) hashPieces() ([]piece.Piece, error) {
	buf := []byte(i.Pieces)

	if len(buf)%bthash.Length != 0 {
		return nil, fmt.Errorf("malformed pieces of length %d", len(buf))
	}

	pieces := make([]piece.Piece, len(buf)/bthash.Length)
	for j := range len(pieces) {
		chunkOffset, chunkEnd := j*bthash.Length, (j+1)*bthash.Length
		pieceOffset, pieceEnd := piece.Bounds(j, i.PieceLength, i.getLength())

		pieces[j] = piece.Piece{
			Index:   j,
			Hash:    bthash.New(buf[chunkOffset:chunkEnd]),
			Offset:  int64(pieceOffset),
			End:     int64(pieceEnd),
			Content: make([]byte, pieceEnd-pieceOffset),
		}
	}

	return pieces, nil
}

// mapFiles transforms each [file] in torrent's metadata into [utils.PathInfo]
// struct with calculated absolute offset and length. These calculated bounds
// may be used to associate downloadable pieces with correct files.
func (i *info) mapFiles() []utils.PathInfo {
	files := make([]utils.PathInfo, len(i.Files))
	offset := 0

	for j, file := range i.Files {
		files[j] = utils.PathInfo{
			Path:   filepath.Join(file.Path...),
			Offset: offset,
			Length: offset + file.Length,
		}

		offset += file.Length
	}

	return files
}

// getLength returns either length value from torrent's metadata or, in case
// of multi-file torrent, sum of all files lengths.
func (i *info) getLength() (length int) {
	if len(i.Files) == 0 {
		return i.Length
	}

	for _, file := range i.Files {
		length += file.Length
	}

	return length
}

// getFiles returns a slice of torrent's files. Single-file torrents also
// may use this function, in which case the file would be given a name after
// the torrent itself and it's bounds would match torrent's size ([0:Length]).
func (i *info) getFiles() []utils.PathInfo {
	if len(i.Files) > 0 {
		return i.mapFiles()
	}

	return []utils.PathInfo{{Path: i.Name, Offset: 0, Length: i.Length}}
}
