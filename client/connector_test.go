package client

import (
	"testing"

	"github.com/sauromates/leech/internal/bitfield"
	"github.com/sauromates/leech/internal/handshake"
	"github.com/sauromates/leech/internal/utils"
	"github.com/stretchr/testify/assert"
)

func TestCompleteHandshake(t *testing.T) {
	type testCase struct {
		serverHandshake []byte
		output          *handshake.Handshake
		shouldFail      bool
	}

	infoHash := utils.BTString{134, 212, 200, 0, 36, 164, 105, 190, 76, 80, 188, 90, 16, 44, 247, 23, 128, 49, 0, 116}
	peerID := utils.BTString{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20}

	tt := map[string]testCase{
		"successful handshake": {
			serverHandshake: []byte{19, 66, 105, 116, 84, 111, 114, 114, 101, 110, 116, 32, 112, 114, 111, 116, 111, 99, 111, 108, 0, 0, 0, 0, 0, 0, 0, 0, 134, 212, 200, 0, 36, 164, 105, 190, 76, 80, 188, 90, 16, 44, 247, 23, 128, 49, 0, 116, 45, 83, 89, 48, 48, 49, 48, 45, 192, 125, 147, 203, 136, 32, 59, 180, 253, 168, 193, 19},
			output: &handshake.Handshake{
				PSTR:     "BitTorrent protocol",
				InfoHash: infoHash,
				PeerID:   utils.BTString{45, 83, 89, 48, 48, 49, 48, 45, 192, 125, 147, 203, 136, 32, 59, 180, 253, 168, 193, 19},
			},
			shouldFail: false,
		},
		"invalid infohash": {
			serverHandshake: []byte{19, 66, 105, 116, 84, 111, 114, 114, 101, 110, 116, 32, 112, 114, 111, 116, 111, 99, 111, 108, 0, 0, 0, 0, 0, 0, 0, 0, 0xde, 0xe8, 0x6a, 0x7f, 0xa6, 0xf2, 0x86, 0xa9, 0xd7, 0x4c, 0x36, 0x20, 0x14, 0x61, 0x6a, 0x0f, 0xf5, 0xe4, 0x84, 0x3d, 45, 83, 89, 48, 48, 49, 48, 45, 192, 125, 147, 203, 136, 32, 59, 180, 253, 168, 193, 19},
			output:          nil,
			shouldFail:      true,
		},
	}

	for _, tc := range tt {
		client, server := createClientAndServer(t)
		server.Write(tc.serverHandshake)

		msg, err := completeHandshake(client, infoHash, peerID)

		if tc.shouldFail {
			assert.NotNil(t, err)
		} else {
			assert.Nil(t, err)
			assert.Equal(t, msg, tc.output)
		}
	}
}

func TestGetBitField(t *testing.T) {
	type testCase struct {
		msg        []byte
		output     bitfield.BitField
		shouldFail bool
	}

	tt := map[string]testCase{
		"valid bitfield": {
			msg:        []byte{0x00, 0x00, 0x00, 0x06, 5, 1, 2, 3, 4, 5},
			output:     bitfield.BitField{1, 2, 3, 4, 5},
			shouldFail: false,
		},
		"invalid message type": {
			msg:        []byte{0x00, 0x00, 0x00, 0x06, 99, 1, 2, 3, 4, 5},
			output:     nil,
			shouldFail: true,
		},
		"keep-alive message": {
			msg:        []byte{0x00, 0x00, 0x00, 0x00},
			output:     nil,
			shouldFail: true,
		},
	}

	for _, tc := range tt {
		client, server := createClientAndServer(t)
		server.Write(tc.msg)

		bf, err := getBitField(client)

		if tc.shouldFail {
			assert.NotNil(t, err)
		} else {
			assert.Nil(t, err)
			assert.Equal(t, bf, tc.output)
		}
	}
}
