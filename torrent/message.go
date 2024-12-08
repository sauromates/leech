package torrent

import (
	"encoding/binary"
	"fmt"
	"io"
)

const (
	// MsgChoke chokes the receiver
	MsgChoke uint8 = 0
	// MsgUnchoke unchokes the receiver
	MsgUnchoke uint8 = 1
	// MsgInterested expresses interest in receiving data
	MsgInterested uint8 = 2
	// MsgNotInterested expresses disinterest in receiving data
	MsgNotInterested uint8 = 3
	// MsgHave alerts the receiver that the sender has downloaded a piece
	MsgHave uint8 = 4
	// MsgBitfield encodes which pieces that the sender has downloaded
	MsgBitfield uint8 = 5
	// MsgRequest requests a block of data from the receiver
	MsgRequest uint8 = 6
	// MsgPiece delivers a block of data to fulfill a request
	MsgPiece uint8 = 7
	// MsgCancel cancels a request
	MsgCancel uint8 = 8
)

type Message struct {
	ID      uint8
	Payload []byte
}

// Serialize serializes a message into a buffer of the form <length><message ID><payload>
//
// Interprets `nil` as a keep-alive message
func (msg *Message) Serialize() []byte {
	if msg == nil {
		return make([]byte, 4)
	}

	length := uint32(len(msg.Payload) + 1) // Calculate length as payload + 1 byte for ID
	content := make([]byte, 4+length)

	binary.BigEndian.PutUint32(content[0:4], length) // Prepend message with length
	content[4] = byte(msg.ID)                        // Put ID after length
	copy(content[5:], msg.Payload)                   // Put payload after ID

	return content
}

// ReadMsg parses a message from a stream. Returns `nil` on keep-alive message
func ReadMsg(r io.Reader) (*Message, error) {
	lengthBuffer := make([]byte, 4)
	if _, err := io.ReadFull(r, lengthBuffer); err != nil {
		return nil, err
	}

	length := binary.BigEndian.Uint32(lengthBuffer)
	if length == 0 {
		return nil, nil // keep-alive message
	}

	content := make([]byte, length)
	if _, err := io.ReadFull(r, content); err != nil {
		return nil, fmt.Errorf("can't read message with length %d (%s)", length, err)
	}

	msg := Message{
		ID:      uint8(content[0]),
		Payload: content[1:],
	}

	return &msg, nil
}

func FormatRequest(index, begin, length int) *Message {
	payload := make([]byte, 12)

	binary.BigEndian.PutUint32(payload[0:4], uint32(index))
	binary.BigEndian.PutUint32(payload[4:8], uint32(begin))
	binary.BigEndian.PutUint32(payload[8:12], uint32(length))

	return &Message{MsgRequest, payload}
}

func FormatHave(index int) *Message {
	payload := make([]byte, 4)
	binary.BigEndian.PutUint32(payload, uint32(index))

	return &Message{MsgHave, payload}
}

func ParsePiece(index int, content []byte, msg *Message) (int, error) {
	if msg.ID != MsgPiece {
		return 0, fmt.Errorf("unexpected message code %d", msg.ID)
	}

	if len(msg.Payload) < 8 {
		return 0, fmt.Errorf("payload is too short: %d < 8", len(msg.Payload))
	}

	parsedIndex := int(binary.BigEndian.Uint32(msg.Payload[0:4]))
	if parsedIndex != index {
		return 0, fmt.Errorf("index mismatch: expected %d, got %d", index, parsedIndex)
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

func ParseHave(msg *Message) (int, error) {
	if msg.ID != MsgHave {
		return 0, fmt.Errorf("unexpected message code %d", msg.ID)
	}

	if len(msg.Payload) != 4 {
		return 0, fmt.Errorf("expected payload of size 4, got %d", len(msg.Payload))
	}

	return int(binary.BigEndian.Uint32(msg.Payload)), nil
}

func (msg *Message) name() string {
	if msg == nil {
		return "KeepAlive"
	}

	switch msg.ID {
	case MsgChoke:
		return "Choke"
	case MsgUnchoke:
		return "Unchoke"
	case MsgInterested:
		return "Interested"
	case MsgNotInterested:
		return "NotInterested"
	case MsgHave:
		return "Have"
	case MsgBitfield:
		return "Bitfield"
	case MsgRequest:
		return "Request"
	case MsgPiece:
		return "Piece"
	case MsgCancel:
		return "Cancel"
	default:
		return fmt.Sprintf("Unknown#%d", msg.ID)
	}
}

func (msg *Message) String() string {
	if msg == nil {
		return msg.name()
	}

	return fmt.Sprintf("%s [%d]", msg.name(), len(msg.Payload))
}
