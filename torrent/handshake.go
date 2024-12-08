package torrent

import (
	"fmt"
	"io"
)

const (
	pstr string = "BitTorrent protocol"
)

// Handshake represents a message exchanged between peers over TCP connection
type Handshake struct {
	PSTR     string
	InfoHash [20]byte
	PeerID   [20]byte
}

// Creates a new handshake with the standard pstr
func NewHandshake(infoHash, peerID [20]byte) *Handshake {
	return &Handshake{
		PSTR:     pstr,
		InfoHash: infoHash,
		PeerID:   peerID,
	}
}

// Serializes handshake message to a buffer
// @todo research some more to ensure how it works
func (message *Handshake) Serialize() []byte {
	buffer := make([]byte, len(message.PSTR)+49)
	buffer[0] = byte(len(message.PSTR))

	curr := 1
	curr += copy(buffer[curr:], message.PSTR)
	curr += copy(buffer[curr:], make([]byte, 8)) // 8 reserved bytes
	curr += copy(buffer[curr:], message.InfoHash[:])
	curr += copy(buffer[curr:], message.PeerID[:])

	return buffer
}

// ReadHandshake parses a handshake from a stream
func ReadHandshake(reader io.Reader) (*Handshake, error) {
	lengthBuffer := make([]byte, 1)
	if _, err := io.ReadFull(reader, lengthBuffer); err != nil {
		return nil, err
	}

	pstrLen := int(lengthBuffer[0])
	if pstrLen == 0 {
		return nil, fmt.Errorf("pstr can't be 0")
	}

	handshakeBuffer := make([]byte, pstrLen+48)
	if _, err := io.ReadFull(reader, handshakeBuffer); err != nil {
		return nil, err
	}

	var infoHash, peerID [20]byte

	// @todo find out how it works
	copy(infoHash[:], handshakeBuffer[pstrLen+8:pstrLen+8+20])
	copy(peerID[:], handshakeBuffer[pstrLen+8+20:])

	message := &Handshake{
		PSTR:     string(handshakeBuffer[0:pstrLen]),
		InfoHash: infoHash,
		PeerID:   peerID,
	}

	return message, nil
}
