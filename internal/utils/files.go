package utils

import (
	"os"
	"path/filepath"
)

// PathInfo holds information about torrent's files with absolute positions.
//
// [PathInfo.Offset] and [PathInfo.Length] may be used to identify piece
// file association in torrent.
type PathInfo struct {
	Path   string // Concatenated file path
	Offset int    // Absolute start position
	Length int    // Absolute end position
}

// FileMap holds data required to save piece to a file.
type FileMap struct {
	Path       string   // Full path to file
	Offset     int64    // Start position to write from
	PieceStart int64    // Offset of the piece to save from
	PieceEnd   int64    // End of the piece to save from
	descriptor *os.File // Holds actual file descriptor
}

// MapPiece converts [PathInfo] into [FileMap] with piece bounds and offset
// relative to file itself and not to the whole torrent.
func (file *PathInfo) MapPiece(offset, pieceStart, pieceEnd int) FileMap {
	return FileMap{
		Path:       file.Path,
		Offset:     int64(offset - file.Offset),
		PieceStart: int64(pieceStart),
		PieceEnd:   int64(pieceEnd),
	}
}

// Open calls [os.OpenFile] and uses the result to set internal pointer to
// opened file.
func (fm *FileMap) Open(name string) error {
	fullPath := filepath.Join(name, fm.Path)

	file, err := os.OpenFile(fullPath, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return err
	}

	fm.descriptor = file

	return err
}

// Close forwards close call to internal file descriptor.
func (fm *FileMap) Close() error { return fm.descriptor.Close() }

// Write implements [io.Writer] interface via [os.File.WriteAt] call.
func (fm FileMap) Write(b []byte) (n int, err error) {
	return fm.descriptor.WriteAt(b, fm.Offset)
}
