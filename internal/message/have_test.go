package message

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateHave(t *testing.T) {
	msg := CreateHave(4)
	expected := &Message{
		ID:      Have,
		Payload: []byte{0x00, 0x00, 0x00, 0x04},
	}

	assert.Equal(t, expected, msg)
}

func TestParseHave(t *testing.T) {
	type testCase struct {
		input      *Message
		output     int
		shouldFail bool
	}

	tests := map[string]testCase{
		"valid message": {
			input:      &Message{Have, []byte{0x00, 0x00, 0x00, 0x04}},
			output:     4,
			shouldFail: false,
		},
		"invalid message type": {
			input:      &Message{Piece, []byte{0x00, 0x00, 0x00, 0x04}},
			output:     0,
			shouldFail: true,
		},
		"too short payload": {
			input:      &Message{Have, []byte{0x00, 0x00, 0x04}},
			output:     0,
			shouldFail: true,
		},
		"too long payload": {
			input:      &Message{Have, []byte{0x00, 0x00, 0x00, 0x00, 0x04}},
			output:     0,
			shouldFail: true,
		},
	}

	for _, test := range tests {
		index, err := test.input.ParseHave()
		if test.shouldFail {
			assert.NotNil(t, err)
		} else {
			assert.Nil(t, err)
		}

		assert.Equal(t, test.output, index)
	}
}
