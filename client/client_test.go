package client

import (
	"net"
	"testing"

	"github.com/sauromates/leech/internal/message"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReadMessage(t *testing.T) {
	client, server := createClientAndServer(t)
	conn := Client{Conn: client}

	msgContents := []byte{
		0x00, 0x00, 0x00, 0x05, // Length
		4,                      // Code `have`
		0x00, 0x00, 0x05, 0x3c, // Payload
	}
	expected := &message.Message{
		ID:      message.Have,
		Payload: []byte{0x00, 0x00, 0x05, 0x3c},
	}

	_, err := server.Write(msgContents)
	require.Nil(t, err)

	msg, err := conn.ReadMessage()

	require.Nil(t, err)
	assert.Equal(t, expected, msg)
}

func TestConfirmHavePiece(t *testing.T) {
	client, server := createClientAndServer(t)
	peer := Client{Conn: client}

	if err := peer.ConfirmHavePiece(1340); err != nil {
		t.Error(err)
	}

	msg := []byte{
		0x00, 0x00, 0x00, 0x05, // Length
		4,                      // Code `have`
		0x00, 0x00, 0x05, 0x3c, // Payload
	}

	assertSuccessfulRead(t, server, msg)
}

func TestRequestUnchoke(t *testing.T) {
	client, server := createClientAndServer(t)
	peer := Client{Conn: client}

	if err := peer.RequestUnchoke(); err != nil {
		t.Error(err)
	}

	msg := []byte{
		0x00, 0x00, 0x00, 0x01,
		1,
	}

	assertSuccessfulRead(t, server, msg)
}

func TestAnnounceInterest(t *testing.T) {
	client, server := createClientAndServer(t)
	peer := Client{Conn: client}

	if err := peer.AnnounceInterest(); err != nil {
		t.Error(err)
	}

	msg := []byte{
		0x00, 0x00, 0x00, 0x01,
		2,
	}

	assertSuccessfulRead(t, server, msg)
}

func TestRequestPiece(t *testing.T) {
	client, server := createClientAndServer(t)
	peer := Client{Conn: client}

	if err := peer.RequestPiece(1, 2, 3); err != nil {
		t.Error(err)
	}

	msg := []byte{
		0x00, 0x00, 0x00, 0x0d, // Length
		6, // Code
		0x00, 0x00, 0x00, 0x01,
		0x00, 0x00, 0x00, 0x02,
		0x00, 0x00, 0x00, 0x03,
	}

	assertSuccessfulRead(t, server, msg)
}

func assertSuccessfulRead(t *testing.T, server net.Conn, expected []byte) {
	buf := make([]byte, len(expected))
	if _, err := server.Read(buf); err != nil {
		t.Error(err)
	}

	assert.Equal(t, expected, buf)
}

// createClientAndServer mocks TCP connection on localhost
func createClientAndServer(t *testing.T) (client, server net.Conn) {
	conn, err := net.Listen("tcp", "127.0.0.1:0")
	require.Nil(t, err)

	// net.Dial does not block, so we need this signalling channel to make sure
	// we don't return before server is ready
	done := make(chan struct{})
	go func() {
		defer conn.Close()

		server, err = conn.Accept()
		require.Nil(t, err)

		done <- struct{}{}
	}()
	client, err = net.Dial("tcp", conn.Addr().String())
	<-done

	return client, server
}
