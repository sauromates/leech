// Package query is responsible for manipulating [DHT queries]
//
// [DHT queries]: https://www.bittorrent.org/beps/bep_0005.html
package query

import (
	"errors"

	"github.com/sauromates/leech/internal/utils"
)

// QueryType represents all possible types of DHT queries.
type QueryType struct {
	val string
}

// Query represents a request sent to nodes in DHT.
type Query struct {
	NodeID    utils.BTString
	QueryType QueryType
}

var (
	Unknown  QueryType = QueryType{""}
	Ping     QueryType = QueryType{"ping"}
	FindNode QueryType = QueryType{"find_node"}
	GetPeers QueryType = QueryType{"get_peers"}
	Announce QueryType = QueryType{"announce_peer"}
)

// ErrUnknownQuery is returned when query response holds unknown type.
var ErrUnknownQuery error = errors.New("[ERROR] unknown query type")

// QueryFromString returns one of the predefined query types or
// [ErrUnknownQuery] in case of unknown value.
func QueryFromString(s string) (QueryType, error) {
	switch s {
	case Ping.val:
		return Ping, nil
	case FindNode.val:
		return FindNode, nil
	case GetPeers.val:
		return GetPeers, nil
	case Announce.val:
		return Announce, nil
	}

	return Unknown, ErrUnknownQuery
}

// String returns query type string representation.
func (q QueryType) String() string { return q.val }
