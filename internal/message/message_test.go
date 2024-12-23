package message

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRead(t *testing.T) {
	type testCase struct {
		input      []byte
		output     *Message
		shouldFail bool
	}

	tests := map[string]testCase{
		"normal message": {
			input:      []byte{0, 0, 0, 5, 4, 1, 2, 3, 4},
			output:     &Message{ID: Have, Payload: []byte{1, 2, 3, 4}},
			shouldFail: false,
		},
		"keep-alive message": {
			input:      []byte{0, 0, 0, 0},
			output:     nil,
			shouldFail: false,
		},
		"too short message": {
			input:      []byte{1, 2, 3},
			output:     nil,
			shouldFail: true,
		},
		"too short buffer": {
			input:      []byte{0, 0, 0, 5, 4, 1, 2},
			output:     nil,
			shouldFail: true,
		},
	}

	for _, test := range tests {
		reader := bytes.NewReader(test.input)
		msg, err := Read(reader)
		if test.shouldFail {
			assert.NotNil(t, err)
		} else {
			assert.Nil(t, err)
		}

		assert.Equal(t, test.output, msg)
	}
}

func TestSerialize(t *testing.T) {
	type testCase struct {
		input  *Message
		output []byte
	}

	tests := map[string]testCase{
		"regular message": {
			input:  &Message{Unchoke, []byte{1, 2, 3, 4}},
			output: []byte{0, 0, 0, 5, 1, 1, 2, 3, 4},
		},
		"have message": {
			input:  &Message{Have, []byte{1, 2, 3, 4}},
			output: []byte{0, 0, 0, 5, 4, 1, 2, 3, 4},
		},
		"request message": {
			input:  &Message{Request, []byte{1, 2, 3, 4}},
			output: []byte{0, 0, 0, 5, 6, 1, 2, 3, 4},
		},
		"keep-alive message": {
			input:  nil,
			output: []byte{0, 0, 0, 0},
		},
	}

	for _, test := range tests {
		buf := test.input.Serialize()
		assert.Equal(t, test.output, buf)
	}
}

func TestString(t *testing.T) {
	type testCase struct {
		input  *Message
		output string
	}

	tests := []testCase{
		{nil, "KeepAlive"},
		{&Message{Choke, []byte{1, 2, 3}}, "Choke [3]"},
		{&Message{Unchoke, []byte{1, 2, 3}}, "Unchoke [3]"},
		{&Message{Interested, []byte{1, 2, 3}}, "Interested [3]"},
		{&Message{NotInterested, []byte{1, 2, 3}}, "NotInterested [3]"},
		{&Message{Have, []byte{1, 2, 3}}, "Have [3]"},
		{&Message{BitField, []byte{1, 2, 3}}, "BitField [3]"},
		{&Message{Request, []byte{1, 2, 3}}, "Request [3]"},
		{&Message{Piece, []byte{1, 2, 3}}, "Piece [3]"},
		{&Message{Cancel, []byte{1, 2, 3}}, "Cancel [3]"},
		{&Message{99, []byte{1, 2, 3}}, "Unknown#99 [3]"},
	}

	for _, test := range tests {
		msg := test.input.String()
		assert.Equal(t, test.output, msg)
	}
}
