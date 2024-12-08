package torrent

import (
	"bytes"
	"fmt"
	"net"
	"time"

	"gihub.com/sauromates/leech/internal/utils"
)

// A Client is a TCP connection with a peer
type Client struct {
	Conn     net.Conn
	IsChoked bool
	BitField utils.BitField
	peer     utils.Peer
	infoHash [20]byte
	peerID   [20]byte
}

func NewClient(peer utils.Peer, infoHash, peerID [20]byte) (*Client, error) {
	conn, err := net.DialTimeout("tcp", peer.String(), 3*time.Second)
	if err != nil {
		return nil, err
	}

	if _, err := connect(conn, infoHash, peerID); err != nil {
		conn.Close()
		return nil, err
	}

	bitfield, err := requestBitfield(conn)
	if err != nil {
		conn.Close()
		return nil, err
	}

	client := &Client{
		Conn:     conn,
		IsChoked: true,
		BitField: bitfield,
		peer:     peer,
		infoHash: infoHash,
		peerID:   peerID,
	}

	return client, nil
}

func (client *Client) Read() (*Message, error) {
	return ReadMsg(client.Conn)
}

func (client *Client) SendRequest(index, begin, length int) error {
	req := FormatRequest(index, begin, length)
	_, err := client.Conn.Write(req.Serialize())

	return err
}

func (client *Client) SendHave(index int) error {
	req := FormatHave(index)
	_, err := client.Conn.Write(req.Serialize())

	return err
}

// @todo return error if message is MsgRequest or MsgHave (they need specific formatting)
func (client *Client) SendMessage(msgType uint8) error {
	msg := Message{ID: msgType}
	_, err := client.Conn.Write(msg.Serialize())

	return err
}

func connect(conn net.Conn, infoHash, peerID [20]byte) (*Handshake, error) {
	conn.SetDeadline(time.Now().Add(3 * time.Second))
	defer conn.SetDeadline(time.Time{}) // Disable the deadline

	req := NewHandshake(infoHash, peerID)
	if _, err := conn.Write(req.Serialize()); err != nil {
		return nil, err
	}

	res, err := ReadHandshake(conn)
	if err != nil {
		return nil, err
	}

	if !bytes.Equal(res.InfoHash[:], infoHash[:]) {
		return nil, fmt.Errorf("torrent info hash mismatch")
	}

	return res, nil
}

func requestBitfield(conn net.Conn) (utils.BitField, error) {
	conn.SetDeadline(time.Now().Add(5 * time.Second))
	defer conn.SetDeadline(time.Time{})

	msg, err := ReadMsg(conn)

	if err != nil {
		return nil, err
	}

	if msg.ID != MsgBitfield {
		return nil, fmt.Errorf("expected bitfield message but got ID %d", msg.ID)
	}

	return msg.Payload, nil
}
