package dht

import (
	"testing"

	"github.com/sauromates/leech/internal/bthash"
	"github.com/stretchr/testify/assert"
)

func TestNewNode(t *testing.T) {
	id, distance := bthash.NewRandom(), 0
	node := NewNode(id, distance, NodeStatusGood)

	assert.Equal(t, id, node.ID)
	assert.Equal(t, int64(distance), node.Distance)
	assert.Equal(t, 0, node.Status)
	assert.Nil(t, node.Conn)
}
