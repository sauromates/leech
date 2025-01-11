package krpc

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMessageFromString(t *testing.T) {
	type testCase struct {
		input  string
		output MessageType
		err    error
	}

	tt := []testCase{
		{"q", TypeQuery, nil},
		{"r", TypeResponse, nil},
		{"e", TypeError, nil},
		{"invalid", TypeUnknown, ErrUnknownMessage},
	}

	for _, tc := range tt {
		msg, err := MessageFromString(tc.input)

		assert.Equal(t, tc.output, msg)
		assert.Equal(t, tc.err, err)
	}
}

func TestString(t *testing.T) {
	type testCase struct {
		input  MessageType
		output string
	}

	tt := []testCase{
		{TypeQuery, "q"},
		{TypeResponse, "r"},
		{TypeError, "e"},
		{TypeUnknown, ""},
	}

	for _, tc := range tt {
		assert.Equal(t, tc.output, tc.input.String())
	}
}
