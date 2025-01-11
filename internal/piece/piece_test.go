package piece

import (
	"bytes"
	"errors"
	"testing"

	"github.com/sauromates/leech/internal/bthash"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	offset, end := 0, 50
	size := 50
	hash := bthash.NewRandom()

	piece := New(0, hash, int64(offset), int64(end))

	assert.Equal(t, int64(offset), piece.Offset)
	assert.Equal(t, int64(end), piece.End)
	assert.Equal(t, hash, piece.Hash)
	assert.Equal(t, size, piece.Size())
	assert.Equal(t, size, len(piece.Content))
}

func TestBounds(t *testing.T) {
	tt := map[string]struct {
		pieceLength   int
		torrentLength int
		index         int
		expectOffset  int
		expectEnd     int
	}{
		"normal piece":      {50, 100, 0, 0, 50},
		"last even piece":   {50, 100, 1, 50, 100},
		"last uneven piece": {13, 100, 7, 91, 100},
	}

	for name, tc := range tt {
		offset, end := Bounds(tc.index, tc.pieceLength, tc.torrentLength)

		assert.Equal(t, tc.expectOffset, offset, name)
		assert.Equal(t, tc.expectEnd, end, name)
	}
}

func TestSize(t *testing.T) {
	piece := Piece{0, bthash.NewRandom(), 0, 10, []byte{}, []byte{}}
	expect := 10

	assert.Equal(t, expect, piece.Size())
}

func TestVerifyHash(t *testing.T) {
	tt := map[string]struct {
		content    []byte
		hash       bthash.Hash
		shouldFail bool
	}{
		"valid contents": {
			content: []byte("The quick brown fox jumps over the lazy dog"),
			hash: bthash.New([]byte{
				47, 212, 225, 198, 122, 45, 40, 252, 237, 132,
				158, 225, 187, 118, 231, 57, 27, 147, 235, 18,
			}),
			shouldFail: false,
		},
		"invalid contents": {
			content: []byte("The quick brown dog jumps over the lazy fox"),
			hash: bthash.New([]byte{
				47, 212, 225, 198, 122, 45, 40, 252, 237, 132,
				158, 225, 187, 118, 231, 57, 27, 147, 235, 18,
			}),
			shouldFail: true,
		},
	}

	for name, tc := range tt {
		piece := Piece{0, tc.hash, 0, 10, tc.content, []byte{}}

		err := piece.VerifyHash()

		if tc.shouldFail {
			assert.Error(t, err, name)
		} else {
			assert.NoError(t, err, name)
		}
	}
}

func TestWriteTo(t *testing.T) {
	tt := map[string]struct {
		piece       *Piece
		from        int64
		to          int64
		makeSection bool
	}{
		"write the whole piece": {
			piece:       New(0, bthash.NewRandom(), 0, 10),
			from:        0,
			to:          10,
			makeSection: true,
		},
		"write part from the beginning": {
			piece:       New(0, bthash.NewRandom(), 0, 10),
			from:        0,
			to:          5,
			makeSection: true,
		},
		"write middle section of piece": {
			piece:       New(0, bthash.NewRandom(), 0, 10),
			from:        5,
			to:          7,
			makeSection: true,
		},
		"write part at the end": {
			piece:       New(0, bthash.NewRandom(), 0, 10),
			from:        7,
			to:          10,
			makeSection: true,
		},
		"write without making a section": {
			piece:       New(0, bthash.NewRandom(), 0, 10),
			from:        0,
			to:          10,
			makeSection: false,
		},
	}

	for name, tc := range tt {
		expectWritten := tc.to - tc.from
		if tc.makeSection {
			tc.piece = tc.piece.Section(tc.from, tc.to)
		}

		var dst bytes.Buffer
		written, err := tc.piece.WriteTo(&dst)

		assert.NoError(t, err, name)
		assert.Equal(t, expectWritten, written, name)
	}
}

func TestWrite(t *testing.T) {
	tt := map[string]struct {
		piece *Piece
		src   []byte
		err   error
	}{
		"valid content size": {
			piece: New(0, bthash.NewRandom(), 0, 10),
			src:   []byte{0, 1, 2, 3, 4, 5},
			err:   nil,
		},
		"too large content": {
			piece: New(0, bthash.NewRandom(), 0, 2),
			src:   []byte{0, 1, 2, 3, 4, 5},
			err:   errors.New("content is larger than piece size: 6"),
		},
	}

	for name, tc := range tt {
		t.Run(name, func(t *testing.T) {
			n, err := tc.piece.Write(tc.src)
			if tc.err != nil {
				assert.Error(t, err, name)
				assert.Equal(t, tc.err, err)
			} else {
				assert.NoError(t, err, name)
				assert.Equal(t, len(tc.src), n)
			}
		})
	}
}

func TestWriteAt(t *testing.T) {
	tt := map[string]struct {
		piece  *Piece
		src    []byte
		offset int64
		err    error
	}{
		"write with valid offset": {
			piece:  New(0, bthash.NewRandom(), 0, 10),
			src:    []byte{0, 1, 2, 3, 4},
			offset: 5,
			err:    nil,
		},
		"offset out of range": {
			piece:  New(0, bthash.NewRandom(), 0, 2),
			src:    []byte{},
			offset: 5,
			err:    errors.New("offset 5 is outside piece bounds"),
		},
	}

	for name, tc := range tt {
		t.Run(name, func(t *testing.T) {
			n, err := tc.piece.WriteAt(tc.src, tc.offset)
			if tc.err != nil {
				assert.Error(t, err)
				assert.Equal(t, tc.err, err)
			} else {
				assert.Equal(t, len(tc.src), n)
				assert.Equal(t, tc.src, tc.piece.Content[tc.offset:])
			}
		})
	}
}

func TestRead(t *testing.T) {
	content := []byte{0, 1, 2, 3, 4, 5}

	piece := New(0, bthash.NewRandom(), 0, 6)
	piece.Content = content

	n, err := piece.Read(content)

	assert.NoError(t, err)
	assert.Equal(t, len(content), n)
}

func TestReadAt(t *testing.T) {
	content := []byte{0, 1, 2, 3, 4, 5}

	piece := New(0, bthash.NewRandom(), 0, 6)
	piece.Content = content

	n, err := piece.ReadAt(content, 0)

	assert.NoError(t, err)
	assert.Equal(t, len(content), n)
}

func TestReader(t *testing.T) {
	content := []byte{0, 1, 2, 3, 4, 5}

	piece := New(0, bthash.NewRandom(), 0, 6)
	piece.Content = content

	assert.Equal(t, bytes.NewReader(content), piece.reader())
}
