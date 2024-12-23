package message

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateRequest(t *testing.T) {
	msg := CreateRequest(4, 567, 4321)
	expected := &Message{
		ID: Request,
		Payload: []byte{
			0x00, 0x00, 0x00, 0x04, // Index
			0x00, 0x00, 0x02, 0x37, // Begin
			0x00, 0x00, 0x10, 0xe1, // Length
		},
	}

	assert.Equal(t, expected, msg)
}

func TestParsePiece(t *testing.T) {
	type testCase struct {
		msg           *Message
		inputIndex    int
		inputBuf      []byte
		outputWritten int
		outputBuf     []byte
		shouldFail    bool
	}

	tt := map[string]testCase{
		"valid piece message": {
			msg: &Message{Piece, []byte{
				0x00, 0x00, 0x00, 0x04, // Index
				0x00, 0x00, 0x00, 0x02, // Begin
				0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff, // Block
			}},
			inputIndex:    4,
			inputBuf:      make([]byte, 10),
			outputWritten: 6,
			outputBuf:     []byte{0x00, 0x00, 0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff, 0x00, 0x00},
			shouldFail:    false,
		},
		"invalid message type": {
			msg:           &Message{Choke, []byte{}},
			inputIndex:    4,
			inputBuf:      make([]byte, 10),
			outputWritten: 0,
			outputBuf:     make([]byte, 10),
			shouldFail:    true,
		},
		"too short payload": {
			msg: &Message{Piece, []byte{
				0x00, 0x00, 0x00, 0x04, // Index
				0x00, 0x00, 0x00, // Malformed offset
			}},
			inputIndex:    4,
			inputBuf:      make([]byte, 10),
			outputWritten: 0,
			outputBuf:     make([]byte, 10),
			shouldFail:    true,
		},
		"invalid piece index": {
			msg: &Message{Piece, []byte{
				0x00, 0x00, 0x00, 0x06, // Index is 6, not 4
				0x00, 0x00, 0x00, 0x02, // Begin
				0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff, // Block
			}},
			inputIndex:    4,
			inputBuf:      make([]byte, 10),
			outputWritten: 0,
			outputBuf:     make([]byte, 10),
			shouldFail:    true,
		},
		"invalid piece offset": {
			msg: &Message{Piece, []byte{
				0x00, 0x00, 0x00, 0x04, // Index
				0x00, 0x00, 0x00, 0x0c, // Begin is 12 > 10
				0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff, // Block
			}},
			inputIndex:    4,
			inputBuf:      make([]byte, 10),
			outputWritten: 0,
			outputBuf:     make([]byte, 10),
			shouldFail:    true,
		},
		"invalid piece payload": {
			msg: &Message{Piece, []byte{
				0x00, 0x00, 0x00, 0x04, // Index
				0x00, 0x00, 0x00, 0x02, // Begin
				// Block is 10 long but begin=2; too long for input buffer
				0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff, 0x0a, 0x0b, 0x0c, 0x0d,
			}},
			inputIndex:    4,
			inputBuf:      make([]byte, 10),
			outputWritten: 0,
			outputBuf:     make([]byte, 10),
			shouldFail:    true,
		},
	}

	for _, tc := range tt {
		written, err := tc.msg.ParsePiece(tc.inputIndex, tc.inputBuf)
		if tc.shouldFail {
			assert.NotNil(t, err)
		} else {
			assert.Nil(t, err)
		}

		assert.Equal(t, tc.outputBuf, tc.inputBuf)
		assert.Equal(t, tc.outputWritten, written)
	}
}
