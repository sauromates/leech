package peers

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUnmarshal(t *testing.T) {
	type testCase struct {
		input      string
		output     []Peer
		shouldFail bool
	}

	tt := map[string]testCase{
		"valid peers string": {
			input: string([]byte{127, 0, 0, 1, 0x00, 0x50, 1, 1, 1, 1, 0x01, 0xbb}),
			output: []Peer{
				{IP: net.IP{127, 0, 0, 1}, Port: 80},
				{IP: net.IP{1, 1, 1, 1}, Port: 443},
			},
			shouldFail: false,
		},
		"not enough bytes in string": {
			input:      string([]byte{127, 0, 0, 1, 0x00}),
			output:     nil,
			shouldFail: true,
		},
	}

	for _, tc := range tt {
		peers, err := Unmarshal([]byte(tc.input))
		if tc.shouldFail {
			assert.NotNil(t, err)
		} else {
			assert.Nil(t, err)
		}

		assert.Equal(t, tc.output, peers)
	}
}

func TestString(t *testing.T) {
	peer := &Peer{IP: net.IP{127, 0, 0, 1}, Port: 8080}
	expected := "127.0.0.1:8080"

	assert.Equal(t, expected, peer.String())
}
