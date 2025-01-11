package message

import (
	"encoding/binary"
	"fmt"

	"github.com/sauromates/leech/internal/piece"
)

// CreateRequest creates a message with code 6 `request`
func CreateRequest(index, begin, length int) *Message {
	payload := make([]byte, 12)

	binary.BigEndian.PutUint32(payload[0:4], uint32(index))
	binary.BigEndian.PutUint32(payload[4:8], uint32(begin))
	binary.BigEndian.PutUint32(payload[8:12], uint32(length))

	return &Message{ID: Request, Payload: payload}
}

// ParsePiece decodes [Piece] message and writes it into given [*piece.Piece].
func (msg *Message) ParsePiece(p *piece.Piece) (int, error) {
	if msg.ID != Piece {
		return 0, fmt.Errorf("unexpected message code %d", msg.ID)
	}

	if len(msg.Payload) < 8 {
		return 0, fmt.Errorf("payload is smaller than 8 bytes")
	}

	msgIndex := int(binary.BigEndian.Uint32(msg.Payload[0:4]))
	if msgIndex != p.Index {
		return 0, fmt.Errorf("piece index mismatch: expected %d but got %d", p.Index, msgIndex)
	}

	begin := int(binary.BigEndian.Uint32(msg.Payload[4:8]))
	if begin >= p.Size() {
		return 0, fmt.Errorf("initial offset is larger than expected (%d over %d)", begin, p.Size())
	}

	payload := msg.Payload[8:]
	if begin+len(payload) > p.Size() {
		return 0, fmt.Errorf("invalid payload length (%d for offset %d with length %d)", len(payload), begin, p.Size())
	}

	return p.WriteAt(payload, int64(begin))
}
