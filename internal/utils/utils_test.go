package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFlattenTaggedStruct(t *testing.T) {
	type testCase struct {
		input  any
		output map[string]any
	}

	type Nested struct {
		Test3 string `json:"test3"`
	}

	tt := map[string]testCase{
		"flat struct with tags": {
			input: struct {
				Test1 string `json:"test1"`
				Test2 string `json:"test2"`
			}{"val1", "val2"},
			output: map[string]any{
				"test1": "val1",
				"test2": "val2",
			},
		},
		"nested struct with tags": {
			input: struct {
				Nested
				Test1 string `json:"test1"`
				Test2 string `json:"test2"`
			}{
				Nested: Nested{"val3"},
				Test1:  "val1",
				Test2:  "val2",
			},
			output: map[string]any{
				"test1": "val1",
				"test2": "val2",
				"test3": "val3",
			},
		},
		"pointer to a nested struct": {
			input: &struct {
				Nested
				Test1 string `json:"test1"`
				Test2 string `json:"test2"`
			}{
				Nested: Nested{"val3"},
				Test1:  "val1",
				Test2:  "val2",
			},
			output: map[string]any{
				"test1": "val1",
				"test2": "val2",
				"test3": "val3",
			},
		},
		"non-struct input": {
			input:  make([]byte, 10),
			output: map[string]any{},
		},
		"empty struct": {
			input:  struct{}{},
			output: map[string]any{},
		},
	}

	for name, tc := range tt {
		assert.Equal(t, tc.output, FlattenTaggedStruct(tc.input, "json"), name)
	}
}

func TestDecodeHEX(t *testing.T) {
	type testCase struct {
		input  string
		output BTString
	}

	tt := map[string]testCase{
		"valid string": {
			input:  "998601FA7B568ABDE138498D12ABA41C1E0CB893",
			output: [20]byte{0x99, 0x86, 0x01, 0xFA, 0x7B, 0x56, 0x8A, 0xBD, 0xE1, 0x38, 0x49, 0x8D, 0x12, 0xAB, 0xA4, 0x1C, 0x1E, 0x0C, 0xB8, 0x93},
		},
		"invalid string": {
			input:  "test",
			output: [20]byte{},
		},
	}

	for name, tc := range tt {
		assert.Equal(t, tc.output, DecodeHEX(tc.input), name)
	}
}
