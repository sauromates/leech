package utils

// PathInfo is processed FileInfo with calculated absolute path, offset
// within the torrent and relative length
type PathInfo struct {
	// Path is a concatenated path from bencoded FileInfo
	Path string
	// Offset is starting point relative to the whole torrent
	Offset int
	// Length is an end position within the whole torrent
	Length int
}
