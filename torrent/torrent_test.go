package torrent

import (
	"crypto/rand"
	"testing"

	"github.com/sauromates/leech/internal/peers"
	"github.com/sauromates/leech/internal/utils"
	"github.com/stretchr/testify/assert"
)

func TestPieceBounds(t *testing.T) {
	type testCase struct {
		torrent    BaseTorrent
		pieceIndex int
		begin      int
		end        int
	}

	tt := map[string]testCase{
		"normal piece": {
			torrent:    createBaseTorrent(50, 100),
			pieceIndex: 0,
			begin:      0,
			end:        50,
		},
		"last piece": {
			torrent:    createBaseTorrent(13, 100),
			pieceIndex: 7,
			begin:      91,
			end:        100,
		},
	}

	for _, tc := range tt {
		begin, end := tc.torrent.PieceBounds(tc.pieceIndex)

		assert.Equal(t, tc.begin, begin)
		assert.Equal(t, tc.end, end)
	}
}

func createBaseTorrent(pieceLength, torrentLength int) BaseTorrent {
	var randPeerID, randInfoHash utils.BTString
	rand.Read(randPeerID[:])
	rand.Read(randInfoHash[:])

	return BaseTorrent{
		Peers:       []peers.Peer{},
		PeerID:      randPeerID,
		InfoHash:    randInfoHash,
		PieceHashes: []utils.BTString{},
		PieceLength: pieceLength,
		Name:        "test",
		Length:      torrentLength,
	}
}
