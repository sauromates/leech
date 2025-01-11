package torrent

import (
	"math"
	"os"
	"testing"

	"github.com/sauromates/leech/internal/bthash"
	"github.com/sauromates/leech/internal/peers"
	"github.com/sauromates/leech/internal/piece"
	"github.com/sauromates/leech/internal/utils"
	"github.com/stretchr/testify/assert"
)

func TestWhichFiles(t *testing.T) {
	type testCase struct {
		torrent       Torrent
		expectedFiles [][]utils.FileMap
		shouldFail    bool
	}

	tt := map[string]testCase{
		"piece size equals file size": {
			torrent: fakeTorrent(50, 100, []utils.PathInfo{
				{Path: "test1", Offset: 0, Length: 50},
				{Path: "test2", Offset: 50, Length: 100},
			}),
			expectedFiles: [][]utils.FileMap{
				{
					{Path: "test1", Offset: 0, PieceStart: 0, PieceEnd: 50},
				},
				{
					{Path: "test2", Offset: 0, PieceStart: 0, PieceEnd: 50},
				},
			},
			shouldFail: false,
		},
		"piece overlaps two files": {
			torrent: fakeTorrent(40, 100, []utils.PathInfo{
				{Path: "test0", Offset: 0, Length: 50},   // 0: [0-40], 1: [40:50] (size 50)
				{Path: "test1", Offset: 50, Length: 80},  // 1: [0:30] (size 30)
				{Path: "test2", Offset: 80, Length: 100}, // 2: [0:20] (size 20)
			}),
			expectedFiles: [][]utils.FileMap{
				{
					{Path: "test0", Offset: 0, PieceStart: 0, PieceEnd: 40},
				},
				{
					{Path: "test0", Offset: 40, PieceStart: 0, PieceEnd: 10},
					{Path: "test1", Offset: 0, PieceStart: 10, PieceEnd: 40},
				},
				{
					{Path: "test2", Offset: 0, PieceStart: 0, PieceEnd: 20},
				},
			},
			shouldFail: false,
		},
		"piece overlaps multiple files": {
			torrent: fakeTorrent(60, 100, []utils.PathInfo{
				{Path: "test0", Offset: 0, Length: 5},
				{Path: "test1", Offset: 5, Length: 10},
				{Path: "test2", Offset: 10, Length: 40},
				{Path: "test3", Offset: 40, Length: 100},
			}),
			expectedFiles: [][]utils.FileMap{
				{
					{Path: "test0", Offset: 0, PieceStart: 0, PieceEnd: 5},
					{Path: "test1", Offset: 0, PieceStart: 5, PieceEnd: 10},
					{Path: "test2", Offset: 0, PieceStart: 10, PieceEnd: 40},
					{Path: "test3", Offset: 0, PieceStart: 40, PieceEnd: 60},
				},
				{
					{Path: "test3", Offset: 20, PieceStart: 0, PieceEnd: 40},
				},
			},
			shouldFail: false,
		},
	}

	for _, tc := range tt {
		for i := range len(tc.torrent.Pieces) {
			files, err := tc.torrent.whichFiles(i)
			if tc.shouldFail {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}

			assert.Equal(t, tc.expectedFiles[i], files)
		}
	}
}

func TestSavePiece(t *testing.T) {
	type expectation struct {
		piece *piece.Piece
		files []utils.PathInfo
	}
	type testCase struct {
		torrent    Torrent
		pieces     []expectation
		total      int64
		shouldFail bool
	}

	tt := map[string]testCase{
		"each piece fits into file": {
			torrent: fakeTorrent(50, 100, []utils.PathInfo{
				{Path: "test0", Offset: 0, Length: 50},
				{Path: "test1", Offset: 50, Length: 100},
			}),
			pieces: []expectation{
				{
					piece: &piece.Piece{Index: 0, Content: make([]byte, 50)},
					files: []utils.PathInfo{
						{Path: "test0", Offset: 0, Length: 50},
					},
				},
				{
					piece: &piece.Piece{Index: 1, Content: make([]byte, 50)},
					files: []utils.PathInfo{
						{Path: "test1", Offset: 0, Length: 50},
					},
				},
			},
			total:      100,
			shouldFail: false,
		},
		"overlapping piece": {
			torrent: fakeTorrent(40, 100, []utils.PathInfo{
				{Path: "test0", Offset: 0, Length: 50},   // 0: [0-40], 1: [40:50] (size 50)
				{Path: "test1", Offset: 50, Length: 80},  // 1: [0:30] (size 30)
				{Path: "test2", Offset: 80, Length: 100}, // 2: [0:20] (size 20)
			}),
			pieces: []expectation{
				{
					piece: &piece.Piece{Index: 0, Content: make([]byte, 40)},
					files: []utils.PathInfo{
						{Path: "test0", Offset: 0, Length: 40},
					},
				},
				{
					piece: &piece.Piece{Index: 1, Content: make([]byte, 40)},
					files: []utils.PathInfo{
						{Path: "test0", Offset: 40, Length: 50},
						{Path: "test1", Offset: 0, Length: 30},
					},
				},
				{
					piece: &piece.Piece{Index: 2, Content: make([]byte, 20)},
					files: []utils.PathInfo{
						{Path: "test2", Offset: 0, Length: 20},
					},
				},
			},
			total:      100,
			shouldFail: false,
		},
	}

	for name, tc := range tt {
		fileSizes := make(map[string]int64, 3)
		for _, expectation := range tc.pieces {
			_, err := tc.torrent.savePiece(expectation.piece)
			if tc.shouldFail {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			for _, file := range expectation.files {
				actualFile, err := os.Stat(file.Path)
				if err != nil {
					assert.Fail(t, err.Error())
				}

				fileSizes[file.Path] = actualFile.Size()

				assert.Equal(t, file.Length, int(actualFile.Size()), name+" case, file "+file.Path)

				os.Remove(file.Path)
			}
		}

		var actualTotal int64 = 0
		for _, size := range fileSizes {
			actualTotal += size
		}

		assert.Equal(t, tc.total, actualTotal, name+" case")
	}
}

// fakeTorrent generates [Torrent] with given parameters.
func fakeTorrent(pieceLength, torrentLength int, files []utils.PathInfo) Torrent {
	pool := make(chan *peers.Peer)

	// Go rounds integers down therefore we need to explicitly convert each
	// value to float64 first and then round division up
	// e.g. 100/40 = 2, math.Ceil(float64(100)/float64(40)) = 3
	pieceCount := math.Ceil(float64(torrentLength) / float64(pieceLength))
	pieces := make([]piece.Piece, int(pieceCount))
	for i := range int(pieceCount) {
		offset, end := piece.Bounds(i, pieceLength, torrentLength)
		pieces[i] = *piece.New(i, bthash.NewRandom(), int64(offset), int64(end))
	}

	return Torrent{
		Name:         "test",
		Length:       torrentLength,
		InfoHash:     bthash.NewRandom(),
		Pieces:       pieces,
		Files:        files,
		Client:       peers.NewFromHost(bthash.NewRandom(), 6881),
		Peers:        pool,
		DownloadPath: "",
	}
}
