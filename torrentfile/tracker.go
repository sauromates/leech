package torrentfile

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/jackpal/bencode-go"
	"github.com/sauromates/leech/internal/peers"
	"github.com/sauromates/leech/internal/utils"
)

// Bencoded response from announcement request
type BencodeTrackerResponse struct {
	// Refresh peers interval in seconds
	Interval int `bencode:"interval"`
	// A blob containing peers' IP addresses and ports
	Peers   string `bencode:"peers"`
	Failure string `bencode:"failure reason"`
}

func (t *TorrentFile) buildTrackerURL(peerID utils.BTString, port uint16) (string, error) {
	baseURL, err := url.Parse(t.Announce)
	if err != nil {
		return "", err
	}

	params := url.Values{
		"info_hash":  []string{string(t.InfoHash[:])},
		"peer_id":    []string{string(peerID[:])},
		"port":       []string{strconv.Itoa(int(port))},
		"uploaded":   []string{"0"},
		"downloaded": []string{"0"},
		"compact":    []string{"1"},
		"left":       []string{strconv.Itoa(t.GetLength())},
	}

	baseURL.RawQuery = params.Encode()

	return baseURL.String(), nil
}

func (t *TorrentFile) requestPeers(peerID utils.BTString, port uint16) ([]peers.Peer, error) {
	url, err := t.buildTrackerURL(peerID, port)
	if err != nil {
		return nil, err
	}

	client := &http.Client{Timeout: 15 * time.Second}
	response, err := client.Get(url)
	if err != nil {
		return nil, err
	}

	defer response.Body.Close()

	tracker := BencodeTrackerResponse{}
	if err := bencode.Unmarshal(response.Body, &tracker); err != nil {
		return nil, err
	}

	if tracker.Failure != "" {
		return nil, fmt.Errorf("request failed: %s", tracker.Failure)
	}

	peers, err := peers.Unmarshal([]byte(tracker.Peers))
	if err != nil {
		return nil, err
	}

	return peers, nil
}
