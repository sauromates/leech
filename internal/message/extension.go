package message

import (
	"bytes"
	"fmt"

	"github.com/jackpal/bencode-go"
)

// NewExtensionHandshake assembles a message for querying peers whether it
// supports some extension. For more details, see [BEP 10]. As of v0.1.0 we
// are requesting only metadata exchange (ut_metadata) extension.
//
// [BEP 10]: https://www.bittorrent.org/beps/bep_0010.html
func NewExtensionHandshake() (*Message, error) {
	m := map[string]any{
		"m": map[string]int{
			"ut_metadata": 1,
		},
	}

	var buf bytes.Buffer
	if err := bencode.Marshal(&buf, m); err != nil {
		return nil, err
	}

	extID := byte(0) // 0 is an ID of handshake message in BEP-10

	// Unlike other messages, extension handshakes reserve 1 byte before
	// actual payload for handshake ID. We're prepending payload with it
	// here to avoid modifying [Serialize()] function.
	payload := append([]byte{extID}, buf.Bytes()...)

	return &Message{Extended, payload}, nil
}

// ParseExtensionHandshake unmarshals bencoded message payload to determine
// whether a peer supports metadata exchange.
func (msg *Message) ParseExtensionHandshake() (bool, error) {
	if msg.ID != Extended {
		return false, fmt.Errorf("unexpected message %s", msg.Name())
	}

	if len(msg.Payload) < 1 {
		return false, fmt.Errorf("invalid payload")
	}

	extID := msg.Payload[0]
	if extID != 0 {
		return false, fmt.Errorf("unexpected extension ID %v", extID)
	}

	// Extensions in format {m: {<extension>: <int>}}
	var extensions map[string]map[string]int
	payload := bytes.NewReader(msg.Payload[1:])
	if err := bencode.Unmarshal(payload, &extensions); err != nil {
		return false, err
	}

	// Check for ut_metadata value in extensions map
	if supports, ok := extensions["m"]["ut_metadata"]; ok && supports == 1 {
		return true, nil
	}

	return false, nil
}
