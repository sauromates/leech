package bthash

import (
	"crypto/rand"
	"errors"
	"io"
	"testing"

	"github.com/pasztorpisti/qs"
)

func TestNewFromHEX(t *testing.T) {
	type testCase struct {
		input    string
		expected Hash
	}

	tt := map[string]testCase{
		"valid hash string": {
			input:    "998601FA7B568ABDE138498D12ABA41C1E0CB893",
			expected: [20]byte{0x99, 0x86, 0x01, 0xFA, 0x7B, 0x56, 0x8A, 0xBD, 0xE1, 0x38, 0x49, 0x8D, 0x12, 0xAB, 0xA4, 0x1C, 0x1E, 0x0C, 0xB8, 0x93},
		},
		"invalid hex chars": {
			input:    "ZZZ",
			expected: Hash{},
		},
		"invalid hash size": {
			input:    "1A",
			expected: Hash{},
		},
	}

	for name, tc := range tt {
		t.Run(name, func(t *testing.T) {
			hash := NewFromHEX(tc.input)
			if hash != tc.expected {
				t.Errorf("%s:\nexpected %s\ngot %d", name, tc.expected, hash)
			}
		})
	}
}

// failingReader is a stub for [io.Reader] that always returns an error
type failingReader struct{}

func (r *failingReader) Read(p []byte) (n int, err error) {
	return n, errors.New("error")
}

func TestNewRandom(t *testing.T) {
	type testCase struct {
		randomizer io.Reader
		shouldFail bool
	}

	tt := map[string]testCase{
		"default randomizer": {
			randomizer: rand.Reader,
			shouldFail: false,
		},
		"failing randomizer": {
			randomizer: &failingReader{},
			shouldFail: true,
		},
	}

	for name, tc := range tt {
		t.Run(name, func(t *testing.T) {
			randomizer = tc.randomizer
			hash := NewRandom()
			isEmpty := hash == (Hash{})

			if tc.shouldFail && !isEmpty {
				t.Errorf("%s should have returned empty hash", name)
			}

			if !tc.shouldFail && isEmpty {
				t.Errorf("%s should have returned valid hash", name)
			}
		})
	}
}

func TestUnmarshalQS(t *testing.T) {
	type testCase struct {
		input      string
		expected   Hash
		shouldFail bool
	}

	tt := map[string]testCase{
		"valid input": {
			input:      "998601FA7B568ABDE138498D12ABA41C1E0CB893",
			expected:   [20]byte{0x99, 0x86, 0x01, 0xFA, 0x7B, 0x56, 0x8A, 0xBD, 0xE1, 0x38, 0x49, 0x8D, 0x12, 0xAB, 0xA4, 0x1C, 0x1E, 0x0C, 0xB8, 0x93},
			shouldFail: false,
		},
		"invalid input": {
			input:      "1A",
			expected:   Hash{},
			shouldFail: true,
		},
	}

	for name, tc := range tt {
		t.Run(name, func(t *testing.T) {
			var query map[string]Hash
			err := qs.Unmarshal(&query, "hash="+tc.input)

			if tc.shouldFail && err == nil {
				t.Errorf("%s was expected to fail", name)
			}

			if query["hash"] != tc.expected {
				t.Errorf("%s:\nexpected: %v\ngot: %v", name, tc.expected, query["hash"])
			}
		})
	}
}

func TestMarshalQS(t *testing.T) {
	type testCase struct {
		hash       map[string]Hash
		query      string
		shouldFail bool
	}

	tt := map[string]testCase{
		"valid bt hash": {
			hash: map[string]Hash{
				"a": {0x99, 0x86, 0x01, 0xFA, 0x7B, 0x56, 0x8A, 0xBD, 0xE1, 0x38, 0x49, 0x8D, 0x12, 0xAB, 0xA4, 0x1C, 0x1E, 0x0C, 0xB8, 0x93},
			},
			query:      "a=%99%86%01%FA%7BV%8A%BD%E18I%8D%12%AB%A4%1C%1E%0C%B8%93",
			shouldFail: false,
		},
		"empty hash": {
			hash: map[string]Hash{
				"a": {},
			},
			query:      "",
			shouldFail: true,
		},
	}

	for name, tc := range tt {
		t.Run(name, func(t *testing.T) {
			query, err := qs.Marshal(tc.hash)

			if tc.shouldFail && err == nil {
				t.Errorf("%s was expected to fail", name)
			}

			if query != tc.query {
				t.Errorf("%s\nexpected: %s\ngot: %s", name, tc.query, query)
			}
		})
	}
}

func TestToHEX(t *testing.T) {
	type testCase struct {
		hash     Hash
		expected string
	}

	tt := map[string]testCase{
		"valid hash": {
			hash:     Hash{0x99, 0x86, 0x01, 0xFA, 0x7B, 0x56, 0x8A, 0xBD, 0xE1, 0x38, 0x49, 0x8D, 0x12, 0xAB, 0xA4, 0x1C, 0x1E, 0x0C, 0xB8, 0x93},
			expected: "998601FA7B568ABDE138498D12ABA41C1E0CB893",
		},
		"empty hash": {
			hash:     Hash{},
			expected: "",
		},
	}

	for name, tc := range tt {
		t.Run(name, func(t *testing.T) {
			s := tc.hash.ToHEX()
			if s != tc.expected {
				t.Errorf("%s\nexpected: %s\ngot: %s", name, tc.expected, s)
			}
		})
	}
}
