package dht

import (
	"math/big"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func init() {
	now = func() time.Time {
		t, _ := time.Parse(time.DateTime, "2025-01-01 00:00:00")
		return t
	}
}

func TestNewBucket(t *testing.T) {
	min, max := big.NewInt(0), MaxRange()
	bucket := NewBucket(*min, *max)

	assert.Equal(t, bucket.MinID, *min)
	assert.Equal(t, bucket.MaxID, *max)
	assert.Empty(t, bucket.Nodes)
	assert.Equal(t, bucket.LastChanged, now())
}

func TestMaxRange(t *testing.T) {
	actual := MaxRange()
	expected, ok := new(big.Int).SetString("1461501637330902918203684832716283019655932542976", 10)
	if !ok {
		t.Fail()
	}

	assert.Equal(t, 0, actual.Cmp(expected))
}

func TestSort(t *testing.T) {
	// Create a new bucket with nodes in random order
	bucket := NewMaxRangeBucket()
	bucket.Nodes = append(bucket.Nodes,
		NewNode(randID(t), 0, NodeStatusGood),         // 0
		NewNode(randID(t), 0, NodeStatusBad),          // 2
		NewNode(randID(t), 0, NodeStatusQuestionable), // 1
		NewNode(randID(t), 0, NodeStatusGood),         // 0
	)

	bucket.Sort()

	// Extract node statuses into a slice after sorting
	statuses := make([]int, len(bucket.Nodes))
	for i, node := range bucket.Nodes {
		statuses[i] = node.Status
	}

	assert.Equal(t, []int{0, 0, 1, 2}, statuses)
}

func TestIsFull(t *testing.T) {
	type testCase struct {
		nodes  []Node
		isFull bool
	}

	tt := map[string]testCase{
		"full bucket": {
			nodes: []Node{
				NewNode(randID(t), 0, NodeStatusGood),
				NewNode(randID(t), 0, NodeStatusGood),
				NewNode(randID(t), 0, NodeStatusGood),
				NewNode(randID(t), 0, NodeStatusGood),
				NewNode(randID(t), 0, NodeStatusGood),
				NewNode(randID(t), 0, NodeStatusGood),
				NewNode(randID(t), 0, NodeStatusGood),
				NewNode(randID(t), 0, NodeStatusGood),
			},
			isFull: true,
		},
		"empty bucket": {
			nodes:  []Node{},
			isFull: false,
		},
	}

	bucket := NewMaxRangeBucket()
	for _, tc := range tt {
		bucket.Nodes = tc.nodes

		assert.Equal(t, tc.isFull, bucket.IsFull())

		bucket.Nodes = []Node{}
	}
}
