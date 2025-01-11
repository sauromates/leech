package dht

import (
	"errors"
	"math/big"
	"sort"
	"time"
)

// BucketSize is default capacity for routing table buckets
const BucketSize int = 8

var (
	ErrBucketIsFull error = errors.New("[ERROR] Bucket is full")
)

type Bucket struct {
	MinID       big.Int
	MaxID       big.Int
	Nodes       []Node
	LastChanged time.Time
}

// NewBucket creates an empty bucket with specified range.
func NewBucket(min, max big.Int) *Bucket {
	return &Bucket{min, max, []Node{}, now()}
}

// NewMaxRangeBucket creates an empty bucket with full range of routing table.
func NewMaxRangeBucket() *Bucket {
	return &Bucket{
		MinID:       *big.NewInt(0),
		MaxID:       *MaxRange(),
		Nodes:       []Node{},
		LastChanged: now(),
	}
}

// MaxRange returns a [big.Int] of 2^160 (full table size).
func MaxRange() *big.Int {
	base := big.NewInt(2)
	pow := big.NewInt(160)

	return new(big.Int).Exp(base, pow, nil)
}

// Sort rearranges nodes in the bucket by their statuses.
func (b *Bucket) Sort() {
	sort.Slice(b.Nodes, func(i, j int) bool {
		return b.Nodes[i].Status <= b.Nodes[j].Status
	})
}

// IsFull determines whether the bucket has reached its full capacity.
func (b *Bucket) IsFull() bool {
	return len(b.Nodes) == BucketSize
}
