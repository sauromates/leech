package message

import (
	"encoding/binary"
	"fmt"
	"io"
)

const (
	Choke         BTMsgID = 0 // Chokes the receiver
	Unchoke       BTMsgID = 1 // Unchokes the receiver
	Interested    BTMsgID = 2 // Expresses interest in receiving data
	NotInterested BTMsgID = 3 // Expresses disinterest in receiving data
	Have          BTMsgID = 4 // Alerts the receiver that the sender has downloaded a piece
	BitField      BTMsgID = 5 // Encodes which pieces the sender has downloaded
	Request       BTMsgID = 6 // Requests a block of data from the receiver
	Piece         BTMsgID = 7 // Delivers a block of data to fulfill a request
	Cancel        BTMsgID = 8 // Cancels a request
)

type BTMsgID uint8

type Message struct {
	ID      BTMsgID
	Payload []byte
}

// Create creates new message with given code as ID
func Create(code BTMsgID) *Message {
	return &Message{ID: code}
}

// Read parses a message from a stream. Returns `nil` on keep-alive message
func Read(r io.Reader) (*Message, error) {
	lenBuf := make([]byte, 4)
	if _, err := io.ReadFull(r, lenBuf); err != nil {
		return nil, err
	}

	length := binary.BigEndian.Uint32(lenBuf)
	if length == 0 {
		return nil, nil // keep-alive message
	}

	payload := make([]byte, length)
	if _, err := io.ReadFull(r, payload); err != nil {
		return nil, err
	}

	return &Message{BTMsgID(payload[0]), payload[1:]}, nil
}

// Serialize serializes a message into a buffer of the form <length><message ID><payload>
// Interprets `nil` as a keep-alive message
func (msg *Message) Serialize() []byte {
	if msg == nil {
		return make([]byte, 4)
	}

	length := uint32(len(msg.Payload) + 1) // Payload + 1 byte for ID
	payload := make([]byte, length+4)

	binary.BigEndian.PutUint32(payload[0:4], length) // Put length at the beginning of message
	payload[4] = byte(msg.ID)                        // Put ID after length
	copy(payload[5:], msg.Payload)                   // Put message payload after ID

	return payload
}

func (msg *Message) String() string {
	if msg == nil {
		return "KeepAlive"
	}

	var name string

	switch msg.ID {
	case Choke:
		name = "Choke"
	case Unchoke:
		name = "Unchoke"
	case Interested:
		name = "Interested"
	case NotInterested:
		name = "NotInterested"
	case Have:
		name = "Have"
	case BitField:
		name = "BitField"
	case Request:
		name = "Request"
	case Piece:
		name = "Piece"
	case Cancel:
		name = "Cancel"
	default:
		name = fmt.Sprintf("Unknown#%d", msg.ID)
	}

	return fmt.Sprintf("%s [%d]", name, len(msg.Payload))
}
