package torrentfile

import (
	"crypto/rand"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	"gihub.com/sauromates/leech/internal/utils"
	"gihub.com/sauromates/leech/torrent"
	"github.com/jackpal/bencode-go"
)

type TorrentFile struct {
	Announce    string
	InfoHash    [20]byte
	PieceHashes [][20]byte
	PieceLength int
	Length      int
	Name        string
	Paths       []utils.FileInfo
}

func (tfile *TorrentFile) Parse() (torrent.Torrent, error) {
	var peerID [20]byte
	if _, err := rand.Read(peerID[:]); err != nil {
		return torrent.Torrent{}, err
	}

	peers, err := tfile.RequestPeers(peerID, uint16(6881))
	if err != nil {
		return torrent.Torrent{}, err
	}

	return torrent.Torrent{
		Peers:       peers,
		PeerID:      peerID,
		InfoHash:    tfile.InfoHash,
		PieceHashes: tfile.PieceHashes,
		PieceLength: tfile.PieceLength,
		Length:      tfile.Length,
		Name:        tfile.Name,
		Paths:       tfile.Paths,
	}, nil
}

func (tfile *TorrentFile) Download(to string) error {
	source, err := tfile.Parse()
	if err != nil {
		return err
	}

	handler := torrent.DownloadWorker{BasePath: filepath.Dir(to)}
	content, err := source.Download(&handler)
	if err != nil {
		return err
	}

	outFile, err := os.Create(to)
	if err != nil {
		return err
	}

	defer outFile.Close()

	if _, err := outFile.Write(content); err != nil {
		return err
	}

	return nil
}

func (torrent *TorrentFile) trackerUrl(peerID [20]byte, port uint16) (string, error) {
	trackerURL, err := url.Parse(torrent.Announce)
	if err != nil {
		return "", err
	}

	queryParams := url.Values{
		"info_hash":  []string{string(torrent.InfoHash[:])},
		"peer_id":    []string{string(peerID[:])},
		"port":       []string{strconv.Itoa(int(port))},
		"uploaded":   []string{"0"},
		"downloaded": []string{"0"},
		"compact":    []string{"1"},
		"left":       []string{strconv.Itoa(torrent.Length)},
	}

	trackerURL.RawQuery = queryParams.Encode()

	return trackerURL.String(), nil
}

func (file *TorrentFile) RequestPeers(peerID [20]byte, port uint16) ([]utils.Peer, error) {
	announceUrl, err := file.trackerUrl(peerID, port)
	if err != nil {
		return nil, err
	}

	httpClient := &http.Client{Timeout: 15 * time.Second}
	res, err := httpClient.Get(announceUrl)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	trackerRes := BencodeTrackerResponse{}
	if err := bencode.Unmarshal(res.Body, &trackerRes); err != nil {
		return nil, err
	}

	return utils.UnmarshalPeers([]byte(trackerRes.Peers))
}

func Open(path string) (TorrentFile, error) {
	file, err := os.Open(path)
	if err != nil {
		return TorrentFile{}, err
	}

	defer file.Close()

	torrent := bencodeTorrent{}
	if err := bencode.Unmarshal(file, &torrent); err != nil {
		return TorrentFile{}, err
	}

	return torrent.createTorrentFile()
}
