// Package dht implements BEP 5 BitTorrent [specification]
//
// [specification]: https://www.bittorrent.org/beps/bep_0005.html
package dht

import (
	"math/big"
	"time"

	"github.com/sauromates/leech/internal/metadata"
	"github.com/sauromates/leech/internal/peers"
)

const bootstrapURL string = "router.utorrent.com:6881"

// now holds reference to [time.Now] function which can be used both in code
// and tests (we can stub it with any func returning time).
var now func() time.Time = time.Now

// RoutingTable is a collection of known nodes split into buckets.
type RoutingTable struct {
	Buckets []Bucket
}

// NewRoutingTable creates a routing table with single bucket with an ID
// space range between 0 and 2^160 (full table size).
func NewRoutingTable() *RoutingTable {
	return &RoutingTable{
		Buckets: []Bucket{*NewMaxRangeBucket()},
	}
}

func FindMetadata() (*metadata.Metadata, error) {
	return &metadata.Metadata{}, nil
}

func FindPeers(pool chan *peers.Peer) {
	//
}

// Insert attempts to put new [Node] to routing table. It will do so by
// searching for appropriate bucket and appending given node to it. If node
// is inserted then bucket contents would be sorted by node statuses.
//
// Returns [ErrBucketIsFull] if no place left for new nodes. It should lead
// to discarding a node.
func (rt *RoutingTable) Insert(node Node) error {
	// Edge case when there's a single bucket - no need for expensive search
	if len(rt.Buckets) == 1 {
		if rt.Buckets[0].IsFull() {
			return ErrBucketIsFull
		}

		rt.Buckets[0].Nodes = append(rt.Buckets[0].Nodes, node)
		rt.Buckets[0].Sort()

		return nil
	}

	id := new(big.Int).SetBytes(node.ID[:])

	i, left, right := 0, 0, len(rt.Buckets)
	for left < right {
		mid := left + (right-left)/2
		midBucket := rt.Buckets[mid]

		if id.Cmp(&midBucket.MinID) >= 0 && id.Cmp(&midBucket.MaxID) <= 0 {
			i = mid
			break
		} else if id.Cmp(&midBucket.MinID) < 0 {
			right = mid - 1
		} else {
			left = mid + 1
		}
	}

	if rt.Buckets[i].IsFull() {
		return ErrBucketIsFull
	}

	rt.Buckets[i].Nodes = append(rt.Buckets[i].Nodes, node)
	rt.Buckets[i].Sort()

	return nil
}

// TotalNodes returns count of all nodes across all table's buckets.
func (rt *RoutingTable) TotalNodes() int {
	total := 0
	for _, bucket := range rt.Buckets {
		total += len(bucket.Nodes)
	}

	return total
}
