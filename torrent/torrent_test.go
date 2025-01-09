package torrent

import (
	"os"
	"testing"

	"github.com/sauromates/leech/internal/bthash"
	"github.com/sauromates/leech/internal/metadata"
	"github.com/sauromates/leech/internal/peers"
	"github.com/sauromates/leech/internal/utils"
	"github.com/sauromates/leech/worker"
	"github.com/stretchr/testify/assert"
)

func TestPieceBounds(t *testing.T) {
	type testCase struct {
		torrent    Torrent
		pieceIndex int
		begin      int
		end        int
	}

	tt := map[string]testCase{
		"normal piece": {
			torrent:    fakeTorrent(50, 100, []metadata.File{}),
			pieceIndex: 0,
			begin:      0,
			end:        50,
		},
		"last piece": {
			torrent:    fakeTorrent(13, 100, []metadata.File{}),
			pieceIndex: 7,
			begin:      91,
			end:        100,
		},
	}

	for _, tc := range tt {
		begin, end := tc.torrent.pieceBounds(tc.pieceIndex)

		assert.Equal(t, tc.begin, begin)
		assert.Equal(t, tc.end, end)
	}
}

func TestWhichFiles(t *testing.T) {
	type testCase struct {
		torrent       Torrent
		expectedFiles []map[string]utils.FileMap
		shouldFail    bool
	}

	tt := map[string]testCase{
		"piece size equals file size": {
			torrent: fakeTorrent(50, 100, []metadata.File{
				{Path: []string{"test1"}, Length: 50},
				{Path: []string{"test2"}, Length: 100},
			}),
			expectedFiles: []map[string]utils.FileMap{
				{
					"test1": {FileName: "test1", FileOffset: 0, PieceStart: 0, PieceEnd: 50},
				},
				{
					"test2": {FileName: "test2", FileOffset: 0, PieceStart: 0, PieceEnd: 50},
				},
			},
			shouldFail: false,
		},
		"piece overlaps two files": {
			torrent: fakeTorrent(40, 100, []metadata.File{
				{Path: []string{"test0"}, Length: 50}, // 0: [0-40], 1: [40:50] (size 50)
				{Path: []string{"test1"}, Length: 30}, // 1: [0:30] (size 30)
				{Path: []string{"test2"}, Length: 20}, // 2: [0:20] (size 20)
			}),
			expectedFiles: []map[string]utils.FileMap{
				{
					"test0": {FileName: "test0", FileOffset: 0, PieceStart: 0, PieceEnd: 40},
				},
				{
					"test0": {FileName: "test0", FileOffset: 40, PieceStart: 0, PieceEnd: 10},
					"test1": {FileName: "test1", FileOffset: 0, PieceStart: 10, PieceEnd: 40},
				},
				{
					"test2": {FileName: "test2", FileOffset: 0, PieceStart: 0, PieceEnd: 20},
				},
			},
			shouldFail: false,
		},
		"piece overlaps multiple files": {
			torrent: fakeTorrent(60, 100, []metadata.File{
				{Path: []string{"test0"}, Length: 5},
				{Path: []string{"test1"}, Length: 5},
				{Path: []string{"test2"}, Length: 30},
				{Path: []string{"test3"}, Length: 60},
			}),
			expectedFiles: []map[string]utils.FileMap{
				{
					"test0": {FileName: "test0", FileOffset: 0, PieceStart: 0, PieceEnd: 5},
					"test1": {FileName: "test1", FileOffset: 0, PieceStart: 5, PieceEnd: 10},
					"test2": {FileName: "test2", FileOffset: 0, PieceStart: 10, PieceEnd: 40},
					"test3": {FileName: "test3", FileOffset: 0, PieceStart: 40, PieceEnd: 60},
				},
				{
					"test3": {FileName: "test3", FileOffset: 20, PieceStart: 0, PieceEnd: 40},
				},
			},
			shouldFail: false,
		},
	}

	for _, tc := range tt {
		for i := range len(tc.expectedFiles) {
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

func TestWrite(t *testing.T) {
	type expectation struct {
		piece *worker.PieceContent
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
			torrent: fakeTorrent(50, 100, []metadata.File{
				{Path: []string{"test0"}, Length: 50},
				{Path: []string{"test1"}, Length: 50},
			}),
			pieces: []expectation{
				{
					piece: &worker.PieceContent{Index: 0, Content: make([]byte, 50)},
					files: []utils.PathInfo{
						{Path: "test0", Offset: 0, Length: 50},
					},
				},
				{
					piece: &worker.PieceContent{Index: 1, Content: make([]byte, 50)},
					files: []utils.PathInfo{
						{Path: "test1", Offset: 0, Length: 50},
					},
				},
			},
			total:      100,
			shouldFail: false,
		},
		"overlapping piece": {
			torrent: fakeTorrent(40, 100, []metadata.File{
				{Path: []string{"test0"}, Length: 50}, // 0: [0-40], 1: [40:50] (size 50)
				{Path: []string{"test1"}, Length: 30}, // 1: [0:30] (size 30)
				{Path: []string{"test2"}, Length: 20}, // 2: [0:20] (size 20)
			}),
			pieces: []expectation{
				{
					piece: &worker.PieceContent{Index: 0, Content: make([]byte, 40)},
					files: []utils.PathInfo{
						{Path: "test0", Offset: 0, Length: 40},
					},
				},
				{
					piece: &worker.PieceContent{Index: 1, Content: make([]byte, 40)},
					files: []utils.PathInfo{
						{Path: "test0", Offset: 40, Length: 50},
						{Path: "test1", Offset: 0, Length: 30},
					},
				},
				{
					piece: &worker.PieceContent{Index: 2, Content: make([]byte, 20)},
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
			err := tc.torrent.write(expectation.piece, nil)
			if tc.shouldFail {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}

			for _, file := range expectation.files {
				actualFile, err := os.Stat(file.Path)
				fileSizes[file.Path] = actualFile.Size()

				assert.Nil(t, err)
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

func fakeTorrent(pieceLength, torrentLength int, files []metadata.File) Torrent {
	pool := make(chan *peers.Peer)

	return Torrent{
		Meta: &metadata.Metadata{
			Announce: "",
			Comment:  "",
			Info: metadata.Info{
				Pieces:      "",
				PieceLength: pieceLength,
				Length:      torrentLength,
				Name:        "test",
				Files:       files,
			},
		},
		ClientID:     bthash.NewRandom(),
		ClientPort:   uint16(6881),
		Peers:        pool,
		DownloadPath: "",
	}
}
