package torrent

import (
	"os"
	"testing"

	"github.com/sauromates/leech/internal/utils"
	"github.com/sauromates/leech/worker"
	"github.com/stretchr/testify/assert"
)

func TestWhichFiles(t *testing.T) {
	type testCase struct {
		torrent       MultiFileTorrent
		expectedFiles []map[string]FileMap
		shouldFail    bool
	}

	tt := map[string]testCase{
		"piece size equals file size": {
			torrent: MultiFileTorrent{
				BaseTorrent: createBaseTorrent(50, 100),
				Paths: []utils.PathInfo{
					{Path: "test1", Offset: 0, Length: 50},
					{Path: "test2", Offset: 50, Length: 100},
				},
			},
			expectedFiles: []map[string]FileMap{
				{
					"test1": {FileOffset: 0, PieceStart: 0, PieceEnd: 50},
				},
				{
					"test2": {FileOffset: 0, PieceStart: 0, PieceEnd: 50},
				},
			},
			shouldFail: false,
		},
		"piece overlaps two files": {
			torrent: MultiFileTorrent{
				BaseTorrent: createBaseTorrent(40, 100),
				Paths: []utils.PathInfo{
					{Path: "test0", Offset: 0, Length: 50},   // 0: [0-40], 1: [40:50] (size 50)
					{Path: "test1", Offset: 50, Length: 80},  // 1: [0:30] (size 30)
					{Path: "test2", Offset: 80, Length: 100}, // 2: [0:20] (size 20)
				},
			},
			expectedFiles: []map[string]FileMap{
				{
					"test0": {FileOffset: 0, PieceStart: 0, PieceEnd: 40},
				},
				{
					"test0": {FileOffset: 40, PieceStart: 0, PieceEnd: 10},
					"test1": {FileOffset: 0, PieceStart: 10, PieceEnd: 40},
				},
				{
					"test2": {FileOffset: 0, PieceStart: 0, PieceEnd: 20},
				},
			},
			shouldFail: false,
		},
		"piece overlaps multiple files": {
			torrent: MultiFileTorrent{
				BaseTorrent: createBaseTorrent(60, 100),
				Paths: []utils.PathInfo{
					{Path: "test0", Offset: 0, Length: 5},
					{Path: "test1", Offset: 5, Length: 10},
					{Path: "test2", Offset: 10, Length: 40},
					{Path: "test3", Offset: 40, Length: 100},
				},
			},
			expectedFiles: []map[string]FileMap{
				{
					"test0": {FileOffset: 0, PieceStart: 0, PieceEnd: 5},
					"test1": {FileOffset: 0, PieceStart: 5, PieceEnd: 10},
					"test2": {FileOffset: 0, PieceStart: 10, PieceEnd: 40},
					"test3": {FileOffset: 0, PieceStart: 40, PieceEnd: 60},
				},
				{
					"test3": {FileOffset: 20, PieceStart: 0, PieceEnd: 40},
				},
			},
			shouldFail: false,
		},
	}

	for _, tc := range tt {
		for i := range len(tc.expectedFiles) {
			files, err := tc.torrent.WhichFiles(i)
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
		piece *worker.TaskResult
		files []utils.PathInfo
	}
	type testCase struct {
		torrent    MultiFileTorrent
		pieces     []expectation
		total      int64
		shouldFail bool
	}

	tt := map[string]testCase{
		"overlapping piece": {
			torrent: MultiFileTorrent{
				BaseTorrent: createBaseTorrent(40, 100),
				Paths: []utils.PathInfo{
					{Path: "test0", Offset: 0, Length: 50},   // 0: [0-40], 1: [40:50] (size 50)
					{Path: "test1", Offset: 50, Length: 80},  // 1: [0:30] (size 30)
					{Path: "test2", Offset: 80, Length: 100}, // 2: [0:20] (size 20)
				},
			},
			pieces: []expectation{
				{
					piece: &worker.TaskResult{Index: 0, Content: make([]byte, 40)},
					files: []utils.PathInfo{
						{Path: "test0", Offset: 0, Length: 40},
					},
				},
				{
					piece: &worker.TaskResult{Index: 1, Content: make([]byte, 40)},
					files: []utils.PathInfo{
						{Path: "test0", Offset: 40, Length: 50},
						{Path: "test1", Offset: 0, Length: 30},
					},
				},
				{
					piece: &worker.TaskResult{Index: 2, Content: make([]byte, 20)},
					files: []utils.PathInfo{
						{Path: "test2", Offset: 0, Length: 20},
					},
				},
			},
			total:      100,
			shouldFail: false,
		},
	}

	for _, tc := range tt {
		fileSizes := make(map[string]int64, 3)
		for _, expectation := range tc.pieces {
			_, err := tc.torrent.Write("", expectation.piece, nil)
			if tc.shouldFail {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}

			for _, file := range expectation.files {
				actualFile, err := os.Stat(file.Path)
				fileSizes[file.Path] = actualFile.Size()

				defer os.Remove(file.Path)

				assert.Nil(t, err)
				// Even though we're writing only a slice to file, its total length
				// is preserved, so writing 10 bytes (from 40 to 50) to 50-bytes sized
				// file should create same 50-bytes sized file, just filled with zeros
				// from 0 to 40
				assert.Equal(t, file.Length, int(actualFile.Size()), "invalid file: "+file.Path)
			}
		}

		var actualTotal int64 = 0
		for _, size := range fileSizes {
			actualTotal += size
		}

		assert.Equal(t, tc.total, actualTotal)
	}
}
