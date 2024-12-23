package peers

import (
	"encoding/binary"
	"fmt"
	"net"
	"strconv"
)

type Peer struct {
	IP   net.IP
	Port uint16
}

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
