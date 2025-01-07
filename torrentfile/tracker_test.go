package torrentfile

import (
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sauromates/leech/internal/bthash"
	"github.com/sauromates/leech/internal/peers"
	"github.com/stretchr/testify/assert"
)

func TestBuildTrackerURL(t *testing.T) {
	torrent := TorrentFile{
		Announce: "http://bttracker.debian.org:6969/announce",
		InfoHash: bthash.Hash{216, 247, 57, 206, 195, 40, 149, 108, 204, 91, 191, 31, 134, 217, 253, 207, 219, 168, 206, 182},
		PieceHashes: []bthash.Hash{
			{49, 50, 51, 52, 53, 54, 55, 56, 57, 48, 97, 98, 99, 100, 101, 102, 103, 104, 105, 106},
			{97, 98, 99, 100, 101, 102, 103, 104, 105, 106, 49, 50, 51, 52, 53, 54, 55, 56, 57, 48},
		},
		PieceLength: 262144,
		Length:      getPointer(351272960),
		Name:        "debian-10.2.0-amd64-netinst.iso",
	}
	peerID := bthash.Hash{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20}
	const port uint16 = 6882

	url, err := torrent.BuildTrackerURL(peerID, port)
	expected := "http://bttracker.debian.org:6969/announce?compact=1&downloaded=0&info_hash=%D8%F79%CE%C3%28%95l%CC%5B%BF%1F%86%D9%FD%CF%DB%A8%CE%B6&left=351272960&peer_id=%01%02%03%04%05%06%07%08%09%0A%0B%0C%0D%0E%0F%10%11%12%13%14&port=6882&uploaded=0"

	if err != nil {
		t.Error(err)
	}

	if url != expected {
		t.Errorf("Invalid URL created: %s", url)
	}
}

func TestRequestPeers(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		content := []byte(
			"d" +
				"8:interval" + "i900e" +
				"5:peers" + "12:" +
				string([]byte{
					192, 0, 2, 123, 0x1A, 0xE1, // 0x1AE1 = 6881
					127, 0, 0, 1, 0x1A, 0xE9, // 0x1AE9 = 6889
				}) + "e")
		res.Write(content)
	}))

	defer server.Close()

	tf := TorrentFile{
		Announce: server.URL,
		InfoHash: bthash.Hash{216, 247, 57, 206, 195, 40, 149, 108, 204, 91, 191, 31, 134, 217, 253, 207, 219, 168, 206, 182},
		PieceHashes: []bthash.Hash{
			{49, 50, 51, 52, 53, 54, 55, 56, 57, 48, 97, 98, 99, 100, 101, 102, 103, 104, 105, 106},
			{97, 98, 99, 100, 101, 102, 103, 104, 105, 106, 49, 50, 51, 52, 53, 54, 55, 56, 57, 48},
		},
		PieceLength: 262144,
		Length:      getPointer(351272960),
		Name:        "debian-10.2.0-amd64-netinst.iso",
	}

	peerID := bthash.Hash{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20}
	const port uint16 = 6882

	expected := []peers.Peer{
		{IP: net.IP{192, 0, 2, 123}, Port: 6881},
		{IP: net.IP{127, 0, 0, 1}, Port: 6889},
	}

	peers, err := tf.RequestPeers(peerID, port)

	assert.Nil(t, err)
	assert.Equal(t, expected, peers)
}

func getPointer(val int) *int {
	return &val
}
