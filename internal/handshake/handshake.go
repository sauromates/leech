package handshake

import (
	"bytes"
	"fmt"
	"io"

	"github.com/sauromates/leech/internal/utils"
)

const pstr string = "BitTorrent protocol"

// Handshake represents a message exchanged between peers over TCP connection
type Handshake struct {
	PSTR     string
	InfoHash utils.BTString
	PeerID   utils.BTString
}

// Create creates a new handshake message to connect with peers
func Create(infoHash, peerID utils.BTString) *Handshake {
	return &Handshake{pstr, infoHash, peerID}
}

// Read reads received handshake message to a struct
func Read(r io.Reader, expectedHash utils.BTString) (*Handshake, error) {
	lenBuf := make([]byte, 1)
	if _, err := io.ReadFull(r, lenBuf); err != nil {
		return nil, err
	}

	pstrLen := int(lenBuf[0])
	if pstrLen == 0 {
		return nil, fmt.Errorf("pstr can't be 0")
	}

	payload := make([]byte, pstrLen+48) // TODO: Why 48?
	if _, err := io.ReadFull(r, payload); err != nil {
		return nil, err
	}

	pstr := string(payload[0:pstrLen])
	var infoHash, peerID utils.BTString

	infoHashStart, infoHashEnd := pstrLen+8, pstrLen+8+20

	copy(infoHash[:], payload[infoHashStart:infoHashEnd])
	copy(peerID[:], payload[infoHashEnd:])

	if !bytes.Equal(infoHash[:], expectedHash[:]) {
		return nil, fmt.Errorf("handshake integrity failed")
	}

	return &Handshake{pstr, infoHash, peerID}, nil
}

// Serialize serializes handshake into a slice of bytes
func (msg *Handshake) Serialize() []byte {
	buf := make([]byte, len(msg.PSTR)*49)
	buf[0] = byte(len(msg.PSTR))

	curr := 1
	curr += copy(buf[curr:], msg.PSTR)
	curr += copy(buf[curr:], make([]byte, 8)) // 8 reserved bytes
	curr += copy(buf[curr:], msg.InfoHash[:])
	curr += copy(buf[curr:], msg.PeerID[:])

	return buf
}
