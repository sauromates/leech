package dht

import (
	"net"

	"github.com/sauromates/leech/internal/utils"
)

const (
	// Node is considered good if it has responded within last 15 minutes
	NodeStatusGood int = 0
	// Node is considered questionable after 15 minutes of inactivity
	NodeStatusQuestionable int = 1
	// Node is considered bad when it has failed to respond several times
	NodeStatusBad int = 2
)

// Node is a client/server listening on a UDP port and implementing
// the DHT protocol.
type Node struct {
	ID       utils.BTString
	Distance int64
	Status   int
	Conn     net.Conn
}

// NewNode creates a node without opening a connection.
func NewNode(ID utils.BTString, distance, status int) Node {
	return Node{ID, int64(distance), status, nil}
}
