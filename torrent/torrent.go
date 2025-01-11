package torrent

import (
	"fmt"

	"github.com/sauromates/leech/internal/bthash"
	"github.com/sauromates/leech/internal/peers"
	"github.com/sauromates/leech/internal/piece"
	"github.com/sauromates/leech/internal/utils"
	"github.com/sauromates/leech/tracker"
	"github.com/sauromates/leech/worker"
	"github.com/schollz/progressbar/v3"
)

// Torrent is a central piece of Leech: it holds all necessary information
// about the target metadata along with download path and client info. It
// should be created as soon as possible, typically within main().
type Torrent struct {
	Name         string
	Length       int
	InfoHash     bthash.Hash
	Client       *peers.Peer
	Pieces       []piece.Piece
	Files        []utils.PathInfo
	Peers        chan *peers.Peer
	Tracker      *tracker.Tracker
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
	queue, err := t.makeQueue()
	if err != nil {
		return err
	}

	done, results := 0, make(chan *piece.Piece)
	bar := progressbar.DefaultBytes(int64(t.Length))

	for done < t.Length {
		select {
		// Write each downloaded piece to files
		case piece := <-results:
			saved, err := t.savePiece(piece)
			if err != nil {
				return err
			}

			done += saved
			bar.Add(saved)
		case peer := <-t.Peers:
			// Run separate goroutine for each new peer
			go func() {
				downloader := worker.New(*peer, t.InfoHash, t.Client.ID)
				err := downloader.Run(queue, results)
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

	if err := t.Tracker.FindPeers(t.Peers); err != nil {
		return err
	}

	return nil
}

// String converts torrent info to default string representation
func (t Torrent) String() string {
	return fmt.Sprintf(
		"Torrent: %s\n---\nTotal size: %.2f Mb\nTotal files: %d\n",
		t.Name,
		float64(t.Length)/(1024*1024),
		len(t.Files),
	)
}

// savePiece searches files associated with given piece and passes them for
// writing. Return value is a number of bytes written.
func (t *Torrent) savePiece(p *piece.Piece) (n int, err error) {
	files, err := t.whichFiles(p.Index)
	if err != nil {
		return n, err
	}

	for _, file := range files {
		err = file.Open(t.DownloadPath)
		if err != nil {
			return n, err
		}

		done, err := p.Section(file.PieceStart, file.PieceEnd).WriteTo(file)
		if err != nil {
			file.Close()
			return n + int(done), err
		}

		n += int(done)
		file.Close()
	}

	if n != p.Size() {
		err = fmt.Errorf("copied %d instead of %d", n, p.Size())
	}

	return n, err
}

// makeQueue puts each piece into a buffered channel.
func (t *Torrent) makeQueue() (chan *piece.Piece, error) {
	queue := make(chan *piece.Piece, len(t.Pieces))
	for _, piece := range t.Pieces {
		queue <- &piece
	}

	return queue, nil
}

// whichFiles determines which files the piece belongs to by an intersection
// of absolute offsets and lengths.
func (t *Torrent) whichFiles(p int) ([]utils.FileMap, error) {
	var files []utils.FileMap
	pieceOffset, pieceEnd := int(t.Pieces[p].Offset), int(t.Pieces[p].End)

	for _, file := range t.Files {
		// Get intersection between piece and file bounds
		intersectStart := max(pieceOffset, file.Offset)
		intersectEnd := min(pieceEnd, file.Length)

		if intersectStart < intersectEnd {
			// Get relative piece bounds
			relStart := intersectStart - pieceOffset
			relEnd := relStart + (intersectEnd - intersectStart)

			files = append(files, file.MapPiece(intersectStart, relStart, relEnd))
		}
	}

	if len(files) == 0 {
		return files, fmt.Errorf("no files found for piece %d", p)
	}

	return files, nil
}
