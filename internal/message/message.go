// Package message enables exchange of BitTorrent messages.
package message

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

const (
	Choke         uint8 = 0  // Chokes the receiver
	Unchoke       uint8 = 1  // Unchokes the receiver
	Interested    uint8 = 2  // Expresses interest in receiving data
	NotInterested uint8 = 3  // Expresses disinterest in receiving data
	Have          uint8 = 4  // Alerts the receiver that the sender has downloaded a piece
	BitField      uint8 = 5  // Encodes which pieces the sender has downloaded
	Request       uint8 = 6  // Requests a block of data from the receiver
	Piece         uint8 = 7  // Delivers a block of data to fulfill a request
	Cancel        uint8 = 8  // Cancels a request
	Extended      uint8 = 20 // Notifies about support of extensions
)

var (
	ErrMessageEmpty   error = errors.New("failed to read empty message")
	ErrInvalidPayload error = errors.New("failed to read message payload")
)

// Message represents request to a peer
type Message struct {
	ID      uint8
	Payload []byte
}

// CreateEmpty creates new message with given code as ID
func CreateEmpty(code uint8) *Message {
	return &Message{ID: code}
}

// Read parses a message from a stream. Returns `nil` on keep-alive message
func Read(r io.Reader) (*Message, error) {
	lenBuf := make([]byte, 4)
	if _, err := io.ReadFull(r, lenBuf); err != nil {
		if err == io.EOF {
			return nil, ErrMessageEmpty
		}
		return nil, err
	}

	length := binary.BigEndian.Uint32(lenBuf)
	if length == 0 {
		return nil, nil // keep-alive message
	}

	payload := make([]byte, length)
	if _, err := io.ReadFull(r, payload); err != nil {
		return nil, ErrInvalidPayload
	}

	return &Message{uint8(payload[0]), payload[1:]}, nil
}

// Serialize serializes a message into a buffer of the form <length><message ID><payload>.
// Interprets `nil` as a keep-alive message
func (msg *Message) Serialize() []byte {
	if msg == nil {
		return make([]byte, 4)
	}

	length := uint32(len(msg.Payload) + 1) // Payload + 1 byte for ID
	payload := make([]byte, 4+length)      // First 4 bytes are for length

	binary.BigEndian.PutUint32(payload[0:4], length) // Put length at the beginning of message
	payload[4] = byte(msg.ID)                        // Put ID after length
	copy(payload[5:], msg.Payload)                   // Put message payload after ID

	return payload
}

// Name returns human-readable string representation of message ID
func (msg *Message) Name() string {
	if msg == nil {
		return "KeepAlive"
	}

	switch msg.ID {
	case Choke:
		return "Choke"
	case Unchoke:
		return "Unchoke"
	case Interested:
		return "Interested"
	case NotInterested:
		return "NotInterested"
	case Have:
		return "Have"
	case BitField:
		return "BitField"
	case Request:
		return "Request"
	case Piece:
		return "Piece"
	case Cancel:
		return "Cancel"
	case Extended:
		return "Extended"
	default:
		return fmt.Sprintf("Unknown#%d", msg.ID)
	}
}

// String transforms message into a string
func (msg *Message) String() string {
	if msg == nil {
		return "KeepAlive"
	}

	return fmt.Sprintf("%s [%d]", msg.Name(), len(msg.Payload))
}
