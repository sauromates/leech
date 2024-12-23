package client

import (
	"net"
	"time"

	"github.com/sauromates/leech/internal/bitfield"
	"github.com/sauromates/leech/internal/message"
	"github.com/sauromates/leech/internal/peers"
	"github.com/sauromates/leech/internal/utils"
)

type Client struct {
	Conn     net.Conn
	IsChoked bool
	BitField bitfield.BitField
	peer     peers.Peer
	infoHash utils.BTString
	peerID   utils.BTString
}

// ReadMessage passes client connection instance as io.Reader to message parser
// and returns parsed struct
func (client *Client) ReadMessage() (*message.Message, error) {
	return message.Read(client.Conn)
}

// ConfirmHavePiece notifies peer about receiving the piece
func (client *Client) ConfirmHavePiece(index int) error {
	return client.sendMessage(message.CreateHave(index))
}

// RequestPiece requests a peer a piece by its index and boundaries
func (client *Client) RequestPiece(index, begin, length int) error {
	return client.sendMessage(message.CreateRequest(index, begin, length))
}

// RequestUnchoke asks a peer to unchoke current client
func (client *Client) RequestUnchoke() error {
	return client.sendMessage(message.Create(message.Unchoke))
}

// AnnounceInterest notifies peer about client being ready to download pieces
func (client *Client) AnnounceInterest() error {
	return client.sendMessage(message.Create(message.Interested))
}

// sendMessage writes any message to a TCP connection pipe
func (client *Client) sendMessage(msg *message.Message) error {
	client.Conn.SetDeadline(time.Now().Add(10 * time.Second))
	defer client.Conn.SetDeadline(time.Time{})

	_, err := client.Conn.Write(msg.Serialize())

	return err
}
