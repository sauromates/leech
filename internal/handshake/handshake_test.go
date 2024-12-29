package handshake

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreate(t *testing.T) {
	infoHash := [20]byte{134, 212, 200, 0, 36, 164, 105, 190, 76, 80, 188, 90, 16, 44, 247, 23, 128, 49, 0, 116}
	peerID := [20]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20}

	actualHandshake := Create(infoHash, peerID)
	expectedHandshake := &Handshake{
		PSTR:     "BitTorrent protocol",
		InfoHash: [20]byte{134, 212, 200, 0, 36, 164, 105, 190, 76, 80, 188, 90, 16, 44, 247, 23, 128, 49, 0, 116},
		PeerID:   [20]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20},
	}

	assert.Equal(t, expectedHandshake, actualHandshake)
}

func TestRead(t *testing.T) {
	type testCase struct {
		input      []byte
		output     *Handshake
		shouldFail bool
	}

	validHash := [20]byte{134, 212, 200, 0, 36, 164, 105, 190, 76, 80, 188, 90, 16, 44, 247, 23, 128, 49, 0, 116}

	tt := map[string]testCase{
		"valid message": {
			input: []byte{19, 66, 105, 116, 84, 111, 114, 114, 101, 110, 116, 32, 112, 114, 111, 116, 111, 99, 111, 108, 0, 0, 0, 0, 0, 0, 0, 0, 134, 212, 200, 0, 36, 164, 105, 190, 76, 80, 188, 90, 16, 44, 247, 23, 128, 49, 0, 116, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20},
			output: &Handshake{
				PSTR:     "BitTorrent protocol",
				InfoHash: validHash,
				PeerID:   [20]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20},
			},
			shouldFail: false,
		},
		"empty message": {
			input:      []byte{},
			output:     nil,
			shouldFail: true,
		},
		"too short message": {
			input:      []byte{19, 66, 105, 116, 84, 111, 114, 114, 101, 110, 116, 32, 112, 114, 111, 116, 111, 99, 111},
			output:     nil,
			shouldFail: true,
		},
		"pstrlen is 0": {
			input:      []byte{0, 0, 0},
			output:     nil,
			shouldFail: true,
		},
	}

	for _, tc := range tt {
		reader := bytes.NewReader(tc.input)
		handshake, err := Read(reader, validHash)

		if tc.shouldFail {
			assert.NotNil(t, err)
		} else {
			assert.Nil(t, err)
		}

		assert.Equal(t, tc.output, handshake)
	}
}

func TestSerialize(t *testing.T) {
	type testCase struct {
		input  *Handshake
		output []byte
	}

	tt := map[string]testCase{
		"serialize message": {
			input: &Handshake{
				PSTR:     "BitTorrent protocol",
				InfoHash: [20]byte{134, 212, 200, 0, 36, 164, 105, 190, 76, 80, 188, 90, 16, 44, 247, 23, 128, 49, 0, 116},
				PeerID:   [20]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20},
			},
			output: []byte{19, 66, 105, 116, 84, 111, 114, 114, 101, 110, 116, 32, 112, 114, 111, 116, 111, 99, 111, 108, 0, 0, 0, 0, 0, 0, 0, 0, 134, 212, 200, 0, 36, 164, 105, 190, 76, 80, 188, 90, 16, 44, 247, 23, 128, 49, 0, 116, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20},
		},
		"different pstr": {
			input: &Handshake{
				PSTR:     "BitTorrent protocol, but cooler?",
				InfoHash: [20]byte{134, 212, 200, 0, 36, 164, 105, 190, 76, 80, 188, 90, 16, 44, 247, 23, 128, 49, 0, 116},
				PeerID:   [20]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20},
			},
			output: []byte{32, 66, 105, 116, 84, 111, 114, 114, 101, 110, 116, 32, 112, 114, 111, 116, 111, 99, 111, 108, 44, 32, 98, 117, 116, 32, 99, 111, 111, 108, 101, 114, 63, 0, 0, 0, 0, 0, 0, 0, 0, 134, 212, 200, 0, 36, 164, 105, 190, 76, 80, 188, 90, 16, 44, 247, 23, 128, 49, 0, 116, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20},
		},
	}

	for _, tc := range tt {
		assert.Equal(t, tc.output, tc.input.Serialize())
	}
}
