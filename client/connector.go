package client

import (
	"bytes"
	"fmt"
	"net"
	"time"

	"github.com/sauromates/leech/internal/bitfield"
	"github.com/sauromates/leech/internal/handshake"
	"github.com/sauromates/leech/internal/message"
	"github.com/sauromates/leech/internal/peers"
	"github.com/sauromates/leech/internal/utils"
)

// Create opens a new TCP connection to a peer
func Create(peer peers.Peer, infoHash, peerID utils.BTString) (*Client, error) {
	conn, err := net.DialTimeout("tcp", peer.String(), 10*time.Second)
	if err != nil {
		return nil, err
	}

	if _, err := completeHandshake(conn, infoHash, peerID); err != nil {
		conn.Close()
		return nil, err
	}

	bitField, err := getBitField(conn)
	if err != nil {
		return nil, err
	}

	client := &Client{
		Conn:     conn,
		IsChoked: true,
		BitField: bitField,
		peer:     peer,
		infoHash: infoHash,
		peerID:   peerID,
	}

	return client, nil
}

// completeHandshake creates and sends new handshake message and reads the
// response into a struct
func completeHandshake(conn net.Conn, infoHash, peerID utils.BTString) (*handshake.Handshake, error) {
	conn.SetDeadline(time.Now().Add(5 * time.Second))
	defer conn.SetDeadline(time.Time{}) // Disables the deadline after handshake

	request := handshake.Create(infoHash, peerID)
	if _, err := conn.Write(request.Serialize()); err != nil {
		return nil, err
	}

	response, err := handshake.Read(conn)
	if err != nil {
		return nil, err
	}

	if !bytes.Equal(response.InfoHash[:], request.InfoHash[:]) {
		return nil, fmt.Errorf("torrent info hash mismatch")
	}

	return response, nil
}

func getBitField(conn net.Conn) (bitfield.BitField, error) {
	conn.SetDeadline(time.Now().Add(10 * time.Second))
	defer conn.SetDeadline(time.Time{})

	msg, err := message.Read(conn)
	if err != nil {
		return nil, err
	}

	if msg == nil {
		return nil, fmt.Errorf("message is empty")
	}

	if msg.ID != message.BitField {
		return nil, fmt.Errorf("received %d instead of bitfield message", msg.ID)
	}

	return msg.Payload, nil
}
