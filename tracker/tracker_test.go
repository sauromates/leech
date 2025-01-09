package tracker

import (
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sauromates/leech/internal/bthash"
	"github.com/sauromates/leech/internal/peers"
	"github.com/stretchr/testify/assert"
)

func TestBuildURL(t *testing.T) {
	tracker := Tracker{
		Announce:   "http://bttracker.debian.org:6969/announce",
		InfoHash:   bthash.Hash{216, 247, 57, 206, 195, 40, 149, 108, 204, 91, 191, 31, 134, 217, 253, 207, 219, 168, 206, 182},
		PeerID:     bthash.Hash{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20},
		Port:       uint16(6882),
		Uploaded:   0,
		Downloaded: 0,
		Compact:    1,
		Left:       351272960,
	}

	url, err := tracker.buildURL()
	expected := "http://bttracker.debian.org:6969/announce?compact=1&downloaded=0&info_hash=%D8%F79%CE%C3%28%95l%CC%5B%BF%1F%86%D9%FD%CF%DB%A8%CE%B6&left=351272960&peer_id=%01%02%03%04%05%06%07%08%09%0A%0B%0C%0D%0E%0F%10%11%12%13%14&port=6882&uploaded=0"

	if err != nil {
		t.Error(err)
	}

	if url != expected {
		t.Errorf("Invalid URL created: %s", url)
	}
}

func TestFindPeers(t *testing.T) {
	handler := func(res http.ResponseWriter, req *http.Request) {
		content := []byte(
			"d" +
				"8:interval" + "i900e" +
				"5:peers" + "12:" +
				string([]byte{
					192, 0, 2, 123, 0x1A, 0xE1, // 0x1AE1 = 6881
					127, 0, 0, 1, 0x1A, 0xE9, // 0x1AE9 = 6889
				}) + "e")
		res.Write(content)
	}
	server := httptest.NewServer(http.HandlerFunc(handler))

	defer server.Close()

	tracker := Tracker{
		// Announce:   "http://bttracker.debian.org:6969/announce",
		Announce:   server.URL,
		InfoHash:   bthash.Hash{216, 247, 57, 206, 195, 40, 149, 108, 204, 91, 191, 31, 134, 217, 253, 207, 219, 168, 206, 182},
		PeerID:     bthash.Hash{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20},
		Port:       uint16(6882),
		Uploaded:   0,
		Downloaded: 0,
		Compact:    1,
		Left:       351272960,
	}

	pool := make(chan *peers.Peer)

	expected := []peers.Peer{
		{IP: net.IP{192, 0, 2, 123}, Port: 6881},
		{IP: net.IP{127, 0, 0, 1}, Port: 6889},
	}

	var actual []peers.Peer
	var err error

	go func() {
		err = tracker.FindPeers(pool)
		close(pool)
	}()

	for peer := range pool {
		actual = append(actual, *peer)
	}

	assert.NoError(t, err)
	assert.Equal(t, expected, actual)
}
