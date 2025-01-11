package krpc

import "errors"

// MessageType represents all possible message types which may be exchanged
// via KRPC protocol
type MessageType struct {
	val string
}

var (
	TypeUnknown  MessageType = MessageType{""}
	TypeQuery    MessageType = MessageType{"q"}
	TypeResponse MessageType = MessageType{"r"}
	TypeError    MessageType = MessageType{"e"}
)

// ErrUnknownMessage is returned when KRPC response contains invalid message ID
var ErrUnknownMessage error = errors.New("[ERROR] Unknown message type")

// MessageFromString returns one of the predefined message types or
// [ErrUnknownMessage] in case of failure
func MessageFromString(s string) (MessageType, error) {
	switch s {
	case TypeQuery.val:
		return TypeQuery, nil
	case TypeResponse.val:
		return TypeResponse, nil
	case TypeError.val:
		return TypeError, nil
	}

	return TypeUnknown, ErrUnknownMessage
}

// String returns message type string representation
func (msg MessageType) String() string { return msg.val }
