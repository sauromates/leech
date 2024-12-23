package message

import (
	"encoding/binary"
	"fmt"
)

// CreateHave creates a message with code 4 (`have`)
func CreateHave(index int) *Message {
	payload := make([]byte, 4)
	binary.BigEndian.PutUint32(payload, uint32(index))

	return &Message{Have, payload}
}

// ParseHave transforms message `HAVE` to an integer value
func (msg *Message) ParseHave() (int, error) {
	if msg.ID != Have {
		return 0, fmt.Errorf("unexpected code %d", msg.ID)
	}

	if len(msg.Payload) != 4 {
		return 0, fmt.Errorf("unexpected payload size %d", len(msg.Payload))
	}

	return int(binary.BigEndian.Uint32(msg.Payload)), nil
}
