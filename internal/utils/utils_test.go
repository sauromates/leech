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
