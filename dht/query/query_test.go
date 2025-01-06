package query

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestQueryFromString(t *testing.T) {
	type testCase struct {
		input  string
		output QueryType
		err    error
	}

	tt := []testCase{
		{"ping", Ping, nil},
		{"find_node", FindNode, nil},
		{"get_peers", GetPeers, nil},
		{"announce_peer", Announce, nil},
		{"invalid_query", Unknown, ErrUnknownQuery},
	}

	for _, tc := range tt {
		query, err := QueryFromString(tc.input)

		assert.Equal(t, tc.output, query)
		assert.Equal(t, tc.err, err)
	}
}

func TestString(t *testing.T) {
	type testCase struct {
		input  QueryType
		output string
	}

	tt := []testCase{
		{Ping, "ping"},
		{FindNode, "find_node"},
		{GetPeers, "get_peers"},
		{Announce, "announce_peer"},
		{Unknown, ""},
	}

	for _, tc := range tt {
		assert.Equal(t, tc.output, tc.input.String())
	}
}
