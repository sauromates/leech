// Package krpc enables support of KRPC messaging protocol
package krpc

// AbstractMessage is a base structure of any KRPC message.
type AbstractMessage struct {
	MessageType   MessageType `krpc:"y"`
	TransactionID uint16      `krpc:"t"`
}

// Error is deserialized from bencoded list ["code", "message"]
type Error struct {
	Code   int
	Reason string
}

// MessageQuery represents incoming query response. Except for default fields
// it also holds `q` for type of query and `a` for response value.
type MessageQuery struct {
	AbstractMessage
	QueryType string      `krpc:"q"`
	Answer    interface{} `krpc:"a"`
}

// MessageResponse represents incoming KRPC response.
type MessageResponse struct {
	AbstractMessage
	Response interface{} `krpc:"r"`
}

// MessageError represents incoming KRPC error response.
type MessageError struct {
	AbstractMessage
	Error Error `krpc:"e"`
}
