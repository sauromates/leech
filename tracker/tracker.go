// Package tracker enables interactions with BitTorrent trackers: requesting
// peers, parsing announce URLs, etc.
package tracker

import (
	"errors"
	"net/http"
	"net/url"
	"time"

	"github.com/jackpal/bencode-go"
	"github.com/pasztorpisti/qs"
	"github.com/sauromates/leech/internal/bthash"
	"github.com/sauromates/leech/internal/peers"
)

// Tracker holds info necessary for making requests to BitTorrent trackers.
//
// All fields except for [Tracker.Announce] are tagged with `qs` tags
// and therefore this struct can be marshaled directly into a query string.
type Tracker struct {
	Announce   string      `qs:"-"`
	InfoHash   bthash.Hash `qs:"info_hash"`
	PeerID     bthash.Hash `qs:"peer_id"`
	Port       uint16      `qs:"port"`
	Uploaded   int         `qs:"uploaded"`
	Downloaded int         `qs:"downloaded"`
	Compact    int         `qs:"compact,omitempty"`
	Left       int         `qs:"left"`
}

// TrackerResponse describes a structure of BitTorrent tracker response
// to peers request. Since the whole response is bencoded this struct provides
// `bencode` tags for quick unmarshaling.
type TrackerResponse struct {
	Interval int    `bencode:"interval" qs:"interval"`
	Peers    string `bencode:"peers" qs:"peers"`
	Failure  string `bencode:"failure reason" qs:"failure reason"`
}

// FindPeers requests tracker for peers information and sends unmarshaled
// values to a given channel.
//
// When using an unbuffered channel make sure to call this function in a
// goroutine otherwise it will block the main thread indefinitely.
func (t *Tracker) FindPeers(pool chan *peers.Peer) error {
	trackerURL, err := t.buildURL()
	if err != nil {
		return err
	}

	client := &http.Client{Timeout: 15 * time.Second}
	response, err := client.Get(trackerURL)
	if err != nil {
		return err
	}

	defer response.Body.Close()

	payload := TrackerResponse{}
	if err := bencode.Unmarshal(response.Body, &payload); err != nil {
		return err
	}

	if payload.Failure != "" {
		return errors.New(payload.Failure)
	}

	peers, err := peers.Unmarshal([]byte(payload.Peers))
	if err != nil {
		return err
	}

	for _, peer := range peers {
		pool <- &peer
	}

	return nil
}

// buildURL marshals tracker fields into a query string and appends it to the
// original announce URL.
func (t *Tracker) buildURL() (string, error) {
	trackerURL, err := url.Parse(t.Announce)
	if err != nil {
		return "", err
	}

	params, err := qs.MarshalValues(t)
	if err != nil {
		return "", err
	}

	trackerURL.RawQuery = params.Encode()

	return trackerURL.String(), nil
}
