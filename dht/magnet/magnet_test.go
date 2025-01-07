package magnet

import (
	"testing"

	"github.com/sauromates/leech/internal/bthash"
	"github.com/stretchr/testify/assert"
)

func TestUnmarshal(t *testing.T) {
	type testCase struct {
		input      string
		output     MagnetLink
		shouldFail bool
	}

	tt := map[string]testCase{
		"valid magnet link": {
			input: "magnet:?xt=urn:btih:998601FA7B568ABDE138498D12ABA41C1E0CB893&tr=http://example.com&dn=test",
			output: MagnetLink{
				InfoHash:   bthash.NewFromString("998601FA7B568ABDE138498D12ABA41C1E0CB893"),
				Name:       "test",
				Length:     0,
				TrackerURL: "http://example.com",
			},
			shouldFail: false,
		},
		"invalid magnet link": {
			input:      "http://example.com?xl=12&xt=test",
			output:     MagnetLink{},
			shouldFail: true,
		},
	}

	for name, tc := range tt {
		var target MagnetLink
		err := Unmarshal(tc.input, &target)

		if tc.shouldFail {
			assert.Error(t, err, name)
		} else {
			assert.NoError(t, err, name)
			assert.Equal(t, tc.output, target, name)
		}
	}
}
