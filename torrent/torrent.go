package torrent

import (
	"errors"
	"fmt"
	"io"
	"path/filepath"

	"github.com/sauromates/leech/internal/bthash"
	"github.com/sauromates/leech/internal/metadata"
	"github.com/sauromates/leech/internal/peers"
	"github.com/sauromates/leech/internal/utils"
	"github.com/sauromates/leech/tracker"
	"github.com/sauromates/leech/worker"
	"github.com/schollz/progressbar/v3"
)

// Torrent is a central piece of Leech: it holds all necessary information
// about the target metadata along with download path and client info. It
// should be created as soon as possible, typically within main().
type Torrent struct {
	Meta         *metadata.Metadata
	ClientID     bthash.Hash
	ClientPort   uint16
	Peers        chan *peers.Peer
	DownloadPath string
}

// Download orchestrates main downloading process and tracks its progress.
//
// It does so by preparing work queue of pieces to download, results channel
// to listen to and a progress tracker implementing [io.Writer]. After all
// preparations torrent would just listen to two channels: results and peers.
//
// Each result will be copied to an associated file and for each peer a
// worker goroutine would be started. Download will listen to mentioned
// channels until the count of downloaded pieces match total piece count.
func (t *Torrent) Download() error {
	pieces, err := t.Meta.Info.HashPieces()
	if err != nil {
		return err
	}

	infoHash, err := t.Meta.Info.Hash()
	if err != nil {
		return err
	}

	// Fill work queue with pieces to download
	queue := make(chan *worker.Piece, len(pieces))
	for i, h := range pieces {
		queue <- &worker.Piece{Index: i, Hash: h, Length: t.pieceSize(i)}
	}

	done, results := 0, make(chan *worker.PieceContent)
	bar := progressbar.DefaultBytes(int64(t.Meta.Info.GetLength()))

	for done < len(pieces) {
		select {
		// Write each downloaded piece to files
		case piece := <-results:
			if err := t.write(piece, bar); err != nil {
				return err
			}

			done++
		case peer := <-t.Peers:
			// Run separate goroutine for each new peer
			job := worker.Create(*peer, infoHash, t.ClientID)
			go func() {
				err := job.Run(queue, results)
				// Return peer to pool in case of any error except ErrConn
				if err != nil && err != worker.ErrConn {
					t.Peers <- peer
				}
			}()
		}
	}

	close(queue)
	close(results)

	return bar.Finish()
}

// FindPeers uses tracker requests and/or DHT to populate peers channel with
// new peers. This function should be called before [Download] to allow the
// former to process incoming peers.
//
// It's important to run it inside a goroutine since populating unbuffered
// channel would block execution.
func (t *Torrent) FindPeers() error {
	// TODO: go dht.FindPeers(t.Peers)

	tracker, err := t.Tracker()
	if err != nil {
		return err
	}

	if err := tracker.FindPeers(t.Peers); err != nil {
		return err
	}

	return nil
}

// Tracker builds tracker info with torrent's metadata. This info may be used
// to request peers from tracker (if one exists, which is generally the case
// only with `.torrent` files).
func (t *Torrent) Tracker() (*tracker.Tracker, error) {
	if t.Meta == nil {
		return nil, errors.New("can't get tracker without torrent metadata")
	}

	infoHash, err := t.Meta.Info.Hash()
	if err != nil {
		return nil, err
	}

	return &tracker.Tracker{
		Announce:   t.Meta.Announce,
		InfoHash:   infoHash,
		PeerID:     t.ClientID,
		Port:       t.ClientPort,
		Uploaded:   0,
		Downloaded: 0,
		Compact:    1,
		Left:       t.Meta.Info.GetLength(),
	}, nil
}

// String converts torrent info to default string representation
func (t Torrent) String() string {
	return fmt.Sprintf(
		"Torrent: %s\n---\nTotal size: %.2f Mb\nTotal files: %d\n",
		t.Meta.Info.Name,
		float64(t.Meta.Info.GetLength())/(1024*1024),
		len(t.Meta.Info.GetFiles()),
	)
}

// write copies received piece into associated files.
//
// Base scenario is writing to a single file, but pieces may overlap files,
// in which case we will split piece by relative offset and length and write
// its parts to multiple associated files
func (t *Torrent) write(p *worker.PieceContent, bar io.Writer) error {
	files, err := t.whichFiles(p.Index)
	if err != nil {
		return err
	}

	for _, file := range files {
		file.Open()

		var dst io.Writer = file
		if bar != nil {
			dst = io.MultiWriter(dst, bar)
		}

		if _, err := p.Save(dst, file.PieceStart, file.PieceEnd); err != nil {
			file.Close()
			return err
		}

		file.Close()
	}

	return nil
}

// pieceBounds calculates where the piece with given index begins
// and ends within the torrent contents
func (t *Torrent) pieceBounds(p int) (begin, end int) {
	begin = p * t.Meta.Info.PieceLength
	end = begin + t.Meta.Info.PieceLength

	maxLength := t.Meta.Info.GetLength()
	if end > maxLength {
		end = maxLength
	}

	return begin, end
}

// pieceSize returns the size of a piece with given index in bytes
func (t *Torrent) pieceSize(p int) int {
	begin, end := t.pieceBounds(p)

	return end - begin
}

// whichFiles determines which files the piece belongs to by an intersection
// of absolute offsets and lengths.
func (t *Torrent) whichFiles(p int) (map[string]utils.FileMap, error) {
	files := make(map[string]utils.FileMap)
	offset, end := t.pieceBounds(p)

	for _, file := range t.Meta.Info.GetFiles() {
		intersectStart := max(offset, file.Offset)
		intersectEnd := min(end, file.Length)

		if intersectStart < intersectEnd {
			relativeStart := intersectStart - offset
			relativeEnd := intersectEnd - intersectStart

			files[file.Path] = utils.FileMap{
				FileName:   filepath.Join(t.DownloadPath, file.Path),
				FileOffset: int64(intersectStart - file.Offset),
				PieceStart: int64(relativeStart),
				PieceEnd:   int64(relativeStart + relativeEnd),
			}
		}
	}

	if len(files) == 0 {
		return files, fmt.Errorf("[ERROR] no files found for piece %d", p)
	}

	return files, nil
}
