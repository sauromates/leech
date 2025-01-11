// Package krpc enables support of KRPC messaging protocol
package krpc

import (
	"bytes"
	"io"

	"github.com/jackpal/bencode-go"
	"github.com/sauromates/leech/dht/query"
	"github.com/sauromates/leech/internal/utils"
)

// messageID is a default value of transaction ID used in KRPC messages.
const messageID string = "aa"

// Message holds common fields for any type of KRPC messages.
type Message struct {
	TransactionID string `bencode:"t"`
	MessageType   string `bencode:"y"`
}

// MessageQuery represents incoming query response. Except for default fields
// it also holds `q` for type of query and `a` for response value.
type MessageQuery struct {
	Message
	QueryType string `bencode:"q"`
	Body      any    `bencode:"a"`
}

// MessageResponse represents incoming KRPC response.
type MessageResponse struct {
	Message
	Response any `bencode:"r"`
}

// MessageError represents incoming KRPC error response.
type MessageError struct {
	Message
	Error []any `bencode:"e"`
}

// NewQueryMessage assembles new KRPC message with [TypeQuery].
func NewQueryMessage(q query.QueryType, body any) (*MessageQuery, error) {
	if q == query.Unknown {
		return nil, query.ErrUnknownQuery
	}

	msg := MessageQuery{
		Message: Message{
			TransactionID: messageID,
			MessageType:   TypeQuery.String(),
		},
		QueryType: q.String(),
		Body:      body,
	}

	return &msg, nil
}

// NewResponseMessage assembles new KRPC message with [TypeResponse].
func NewResponseMessage(body any) *MessageResponse {
	return &MessageResponse{
		Message: Message{
			TransactionID: messageID,
			MessageType:   TypeResponse.String(),
		},
		Response: body,
	}
}

// NewResponseMessage assembles new KRPC message with [TypeError].
func NewErrorMessage(body []any) *MessageError {
	return &MessageError{
		Message: Message{
			TransactionID: messageID,
			MessageType:   TypeError.String(),
		},
		Error: body,
	}
}

// Serialize marshals KRPC message into bencode dictionary.
func Serialize(msg any) ([]byte, error) {
	flatMsg := utils.FlattenTaggedStruct(msg, "bencode")

	var buf bytes.Buffer
	if err := bencode.Marshal(&buf, flatMsg); err != nil {
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
