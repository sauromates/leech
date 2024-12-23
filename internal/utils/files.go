package utils

type FileInfo struct {
	Length int      `bencode:"length"`
	Path   []string `bencode:"path"`
}
