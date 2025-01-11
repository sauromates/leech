package peers

import (
	"encoding/binary"
	"fmt"
	"net"
	"strconv"

	"github.com/sauromates/leech/internal/bthash"
)

// Peer is a participant of P2P network. In context of DHT, peer is also a
// node.
//
// The application itself is also considered a peer.
type Peer struct {
	ID   bthash.Hash
	IP   net.IP
	Port uint16
}

// New allows to create a peer with IP passed as string.
func New(ip string, port uint16, id bthash.Hash) *Peer {
	return &Peer{ID: id, IP: net.ParseIP(ip), Port: port}
}

// NewFromHost makes peer from the host system.
func NewFromHost(port uint16) *Peer {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return nil
	}

	defer conn.Close()

	return New(conn.LocalAddr().String(), port, bthash.NewRandom())
}

// Unmarshal decodes raw bytes into a slice of [Peer] structs.
func Unmarshal(raw []byte) ([]Peer, error) {
	const size int = 6 // 4 for IP, 2 for port
	if len(raw)%size != 0 {
		return nil, fmt.Errorf("received malformed peers data of length %d", len(raw))
	}

	count := len(raw) / size
	peers := make([]Peer, count)
	for i := range count {
		offset := i * size

		peers[i].IP = net.IP(raw[offset : offset+4])
		peers[i].Port = binary.BigEndian.Uint16([]byte(raw[offset+4 : offset+6]))
	}

	return peers, nil
}

func (p *Peer) String() string {
	return net.JoinHostPort(p.IP.String(), strconv.Itoa(int(p.Port)))
}
