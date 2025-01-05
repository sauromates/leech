package utils

import (
	"io"
	"os"
)

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

// FileMap holds metadata required to write piece into a specific place
// inside a file. For convenience it implements both [io.Writer] and
// [io.WriterAt] interfaces and may be used as [os.File] to open and close
// it at will.
type FileMap struct {
	// FileName is a full path to a file which can be used to open/close it
	FileName string
	// FileOffset is a start position for underlying [io.WriterAt]
	FileOffset int64
	// PieceStart is a lower bound for a piece chunk to write
	PieceStart int64
	// PieceEnd is an upper bound for a piece chunk to write
	PieceEnd int64
	// Descriptor allows to use FileMap as an instance of [os.File]
	Descriptor *os.File
}

// Open uses [os.OpenFile] call to open the file with default settings. In
// case of success FileMap will receive a pointer to a file descriptor.
func (fm *FileMap) Open() (io.Writer, error) {
	file, err := os.OpenFile(fm.FileName, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}

	fm.Descriptor = file

	return file, nil
}

// Close calls [os.File.Close]
func (fm *FileMap) Close() error {
	return fm.Descriptor.Close()
}

// Write calls underlying [io.Writer]
func (fm FileMap) Write(b []byte) (n int, err error) {
	return fm.writer().Write(b)
}

// WriteAt calls underlying [io.WriterAt]
func (fm FileMap) WriteAt(b []byte, off int64) (n int, err error) {
	return fm.writer().WriteAt(b, off)
}

// writer returns an instance of [*io.OffsetWriter] which is used in all
// implementations of [io.Writer] and [io.WriterAt] when dealing with file.
func (fm *FileMap) writer() *io.OffsetWriter {
	return io.NewOffsetWriter(fm.Descriptor, fm.FileOffset)
}
