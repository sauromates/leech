package utils

import (
	"encoding/binary"
	"fmt"
	"net"
	"strconv"
)

// Peer encodes connection information for a peer
type Peer struct {
	IP   net.IP
	Port uint16
}

func UnmarshalPeers(blob []byte) ([]Peer, error) {
	peerSize := 6
	peerCount := len(blob) / peerSize

	if len(blob)%peerSize != 0 {
		return nil, fmt.Errorf("received malformed peers")
	}

	peers := make([]Peer, peerCount)
	for i := range peerCount {
		offset := i * peerSize
		peers[i].IP = net.IP(blob[offset : offset+4])
		peers[i].Port = binary.BigEndian.Uint16([]byte(blob[offset+4 : offset+6]))
	}

	return peers, nil
}

func (peer Peer) String() string {
	return net.JoinHostPort(peer.IP.String(), strconv.Itoa(int(peer.Port)))
}
