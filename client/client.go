package client

import (
	"net"
	"time"

	"github.com/sauromates/leech/internal/bitfield"
	"github.com/sauromates/leech/internal/message"
	"github.com/sauromates/leech/internal/peers"
)

// Client represents a TCP connection with a peer
type Client struct {
	Conn     net.Conn
	IsChoked bool
	BitField bitfield.BitField
	Peer     peers.Peer
}

// Read passes client connection instance as io.Reader to message parser
// and returns parsed struct
func (client *Client) Read() (*message.Message, error) {
	client.Conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	defer client.Conn.SetReadDeadline(time.Time{})

	return message.Read(client.Conn)
}

// Write sends any message over a TCP connection
func (client *Client) Write(msg *message.Message) error {
	client.Conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
	defer client.Conn.SetWriteDeadline(time.Time{})

	_, err := client.Conn.Write(msg.Serialize())

	return err
}

// ConfirmHavePiece notifies peer about receiving the piece
func (client *Client) ConfirmHavePiece(index int) error {
	return client.Write(message.CreateHave(index))
}

// RequestPiece requests a peer for a piece by its index, offset and length
func (client *Client) RequestPiece(index, begin, length int) error {
	return client.Write(message.CreateRequest(index, begin, length))
}

// Unchoke asks a peer to unchoke current client
func (client *Client) Unchoke() error {
	return client.Write(message.CreateEmpty(message.Unchoke))
}

// AnnounceInterest notifies peer about client being ready to download pieces
func (client *Client) AnnounceInterest() error {
	return client.Write(message.CreateEmpty(message.Interested))
}
