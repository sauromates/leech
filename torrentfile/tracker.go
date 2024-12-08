package torrentfile

// Bencoded response from announcement request
type BencodeTrackerResponse struct {
	// Refresh peers interval in seconds
	Interval int `bencode:"interval"`
	// A blob containing peers' IP addresses and ports
	Peers string `bencode:"peers"`
}
