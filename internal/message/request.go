package message

import (
	"encoding/binary"
	"fmt"
)

// CreateRequest creates a message with code 6 `request`
func CreateRequest(index, begin, length int) *Message {
	payload := make([]byte, 12)

	binary.BigEndian.PutUint32(payload[0:4], uint32(index))
	binary.BigEndian.PutUint32(payload[4:8], uint32(begin))
	binary.BigEndian.PutUint32(payload[8:12], uint32(length))

	return &Message{ID: Request, Payload: payload}
}

// ParsePiece verifies incoming message and returns length of downloaded piece
func (msg *Message) ParsePiece(index int, content []byte) (int, error) {
	if msg.ID != Piece {
		return 0, fmt.Errorf("unexpected message code %d", msg.ID)
	}

	if len(msg.Payload) < 8 {
		return 0, fmt.Errorf("payload is smaller than 8 bytes")
	}

	msgIndex := int(binary.BigEndian.Uint32(msg.Payload[0:4]))
	if msgIndex != index {
		return 0, fmt.Errorf("piece index mismatch: expected %d but got %d", index, msgIndex)
	}

	begin := int(binary.BigEndian.Uint32(msg.Payload[4:8]))
	if begin >= len(content) {
		return 0, fmt.Errorf("initial offset is larger than expected (%d over %d)", begin, len(content))
	}

	payload := msg.Payload[8:]
	if begin+len(payload) > len(content) {
		return 0, fmt.Errorf("invalid payload length (%d for offset %d with length %d)", len(payload), begin, len(content))
	}

	copy(content[begin:], payload)

	return len(payload), nil
}
