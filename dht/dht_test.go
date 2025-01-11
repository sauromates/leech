package dht

import (
	"math/big"
	"testing"

	"github.com/sauromates/leech/internal/bthash"
	"github.com/stretchr/testify/assert"
)

func TestNewRoutingTable(t *testing.T) {
	table := NewRoutingTable()

	assert.Len(t, table.Buckets, 1)

	bucket := table.Buckets[0]

	assert.Equal(t, bucket.MinID, *big.NewInt(0))
	assert.Equal(t, bucket.MaxID, *MaxRange())
	assert.Empty(t, bucket.Nodes)
	assert.Equal(t, bucket.LastChanged, now())
}

func TestInsert(t *testing.T) {
	type testCase struct {
		rt         *RoutingTable
		node       Node
		shouldFail bool
	}

	tt := map[string]testCase{
		"table with single empty bucket": {
			rt:         NewRoutingTable(),
			node:       NewNode(bthash.NewRandom(), 0, NodeStatusQuestionable),
			shouldFail: false,
		},
		"table with single full bucket": {
			rt: &RoutingTable{
				Buckets: []Bucket{
					{
						MinID: *big.NewInt(0),
						MaxID: *MaxRange(),
						Nodes: []Node{
							NewNode(bthash.NewRandom(), 0, NodeStatusGood),
							NewNode(bthash.NewRandom(), 0, NodeStatusGood),
							NewNode(bthash.NewRandom(), 0, NodeStatusGood),
							NewNode(bthash.NewRandom(), 0, NodeStatusGood),
							NewNode(bthash.NewRandom(), 0, NodeStatusGood),
							NewNode(bthash.NewRandom(), 0, NodeStatusGood),
							NewNode(bthash.NewRandom(), 0, NodeStatusGood),
							NewNode(bthash.NewRandom(), 0, NodeStatusGood),
						},
						LastChanged: now(),
					},
				},
			},
			node:       NewNode(bthash.NewRandom(), 0, NodeStatusQuestionable),
			shouldFail: true,
		},
		"table with available bucket": {
			rt: &RoutingTable{[]Bucket{
				{
					MinID:       *big.NewInt(0),
					MaxID:       *new(big.Int).Exp(big.NewInt(2), big.NewInt(80), nil), // 2^80
					Nodes:       []Node{},
					LastChanged: now(),
				},
				{
					MinID:       *new(big.Int).Exp(big.NewInt(2), big.NewInt(80), nil),
					MaxID:       *MaxRange(),
					Nodes:       []Node{},
					LastChanged: now(),
				},
			}},
			node:       NewNode(bthash.NewRandom(), 0, NodeStatusQuestionable),
			shouldFail: false,
		},
		"table with full buckets": {
			rt: &RoutingTable{[]Bucket{
				{
					MinID: *big.NewInt(0),
					MaxID: *new(big.Int).Exp(big.NewInt(2), big.NewInt(80), nil), // 2^80
					Nodes: []Node{
						NewNode(bthash.NewRandom(), 0, NodeStatusGood),
						NewNode(bthash.NewRandom(), 0, NodeStatusGood),
						NewNode(bthash.NewRandom(), 0, NodeStatusGood),
						NewNode(bthash.NewRandom(), 0, NodeStatusGood),
						NewNode(bthash.NewRandom(), 0, NodeStatusGood),
						NewNode(bthash.NewRandom(), 0, NodeStatusGood),
						NewNode(bthash.NewRandom(), 0, NodeStatusGood),
						NewNode(bthash.NewRandom(), 0, NodeStatusGood),
					},
					LastChanged: now(),
				},
				{
					MinID: *new(big.Int).Exp(big.NewInt(2), big.NewInt(80), nil),
					MaxID: *MaxRange(),
					Nodes: []Node{
						NewNode(bthash.NewRandom(), 0, NodeStatusGood),
						NewNode(bthash.NewRandom(), 0, NodeStatusGood),
						NewNode(bthash.NewRandom(), 0, NodeStatusGood),
						NewNode(bthash.NewRandom(), 0, NodeStatusGood),
						NewNode(bthash.NewRandom(), 0, NodeStatusGood),
						NewNode(bthash.NewRandom(), 0, NodeStatusGood),
						NewNode(bthash.NewRandom(), 0, NodeStatusGood),
						NewNode(bthash.NewRandom(), 0, NodeStatusGood),
					},
					LastChanged: now(),
				},
			}},
			node:       NewNode(bthash.NewRandom(), 0, NodeStatusQuestionable),
			shouldFail: true,
		},
	}

	for name, test := range tt {
		wasCount := test.rt.TotalNodes()
		err := test.rt.Insert(test.node)
		if test.shouldFail {
			assert.NotNil(t, err, name)
		} else {
			assert.Nil(t, err, name)
			assert.Equal(t, wasCount+1, test.rt.TotalNodes(), name)
		}
	}
}
