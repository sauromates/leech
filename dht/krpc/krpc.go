// Package krpc enables support of KRPC messaging protocol
package krpc

import (
	"bytes"
	"io"

	"github.com/jackpal/bencode-go"
	"github.com/sauromates/leech/dht/query"
)

// Error is deserialized from bencoded list ["code", "message"]
type Error struct {
	Code   int
	Reason string
}

// MessageQuery represents incoming query response. Except for default fields
// it also holds `q` for type of query and `a` for response value.
type MessageQuery struct {
	TransactionID string      `bencode:"t"`
	MessageType   string      `bencode:"y"`
	QueryType     string      `bencode:"q"`
	Body          interface{} `bencode:"a"`
}

// MessageResponse represents incoming KRPC response.
type MessageResponse struct {
	TransactionID string      `bencode:"t"`
	MessageType   string      `bencode:"y"`
	Response      interface{} `bencode:"r"`
}

// MessageError represents incoming KRPC error response.
type MessageError struct {
	TransactionID string `bencode:"t"`
	MessageType   string `bencode:"y"`
	Error         []any  `bencode:"e"`
}

// NewQueryMessage assembles new KRPC message with [TypeQuery].
func NewQueryMessage(q query.QueryType, body interface{}) (*MessageQuery, error) {
	if q == query.Unknown {
		return nil, query.ErrUnknownQuery
	}

	msg := MessageQuery{
		TransactionID: "aa",
		MessageType:   TypeQuery.String(),
		QueryType:     q.String(),
		Body:          body,
	}

	return &msg, nil
}

func NewResponseMessage(body interface{}) *MessageResponse {
	return &MessageResponse{
		TransactionID: "aa",
		MessageType:   TypeResponse.String(),
		Response:      body,
	}
}

func NewErrorMessage(body []any) *MessageError {
	return &MessageError{
		TransactionID: "aa",
		MessageType:   TypeError.String(),
		Error:         body,
	}
}

// Serialize marshals KRPC message into bencode dictionary.
func Serialize(msg interface{}) ([]byte, error) {
	var buf bytes.Buffer
	if err := bencode.Marshal(&buf, msg); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// Deserialize decodes bencode dictionary into a generic message.
func Deserialize[T any](msg io.Reader) (T, error) {
	var response T
	if err := bencode.Unmarshal(msg, &response); err != nil {
		return response, err
	}

	return response, nil
}
