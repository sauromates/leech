package main

import (
	"fmt"
	"log"
	"os"

	"github.com/sauromates/leech/dht"
	"github.com/sauromates/leech/internal/bthash"
	"github.com/sauromates/leech/internal/metadata"
	"github.com/sauromates/leech/internal/peers"
	"github.com/sauromates/leech/magnetlink"
)

var (
	clientID bthash.Hash = bthash.NewRandom()
	port     uint16      = 6881
	pool     chan *peers.Peer
)

func init() {
	if err := configureLogs("leech.log"); err != nil {
		exit(err)
	}

	pool = make(chan *peers.Peer)
}

func main() {
	source := os.Args[1]

	defer close(pool)

	meta, err := getMetadata(source)
	if err != nil {
		exit(err)
	}

	torrent, err := meta.NewTorrent(peers.NewFromHost(port))
	if err != nil {
		exit(err)
	}

	dir, err := createDownloadDir(torrent.Name)
	if err != nil {
		exit(err)
	}

	torrent.DownloadPath = dir
	torrent.Peers = pool

	fmt.Printf("Downloading\n---\n%s\n", torrent)

	// Spin up peer discovery in a separate goroutine
	go torrent.FindPeers()

	if err := torrent.Download(); err != nil {
		exit(err)
	}

	printResultDetails(dir)
}

// getMetadata issues a search for torrent metadata based on input's type:
// magnet link input will lead to search in DHT for a peer holding the info
// while `.torrent` file inputs are decoded via [torrentfile] package.
func getMetadata(src string) (*metadata.Metadata, error) {
	if magnetlink.Check(src) {
		return dht.FindMetadata()
	}

	file, err := os.Open(src)
	if err != nil {
		return nil, err
	}

	defer file.Close()

	return metadata.Parse(file)
}

// exit prints error and returns non-zero code
func exit(err error) {
	fmt.Println(err)
	os.Exit(1)
}

// configureLogs sets default log output to a file with given path
func configureLogs(path string) error {
	logFile, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	defer logFile.Close()

	log.SetOutput(logFile)
	log.SetFlags(log.Lshortfile | log.LstdFlags)

	return nil
}

// createDownloadDir creates a directory to store downloaded files
func createDownloadDir(torrentName string) (string, error) {
	outputPath, err := os.Getwd()
	if err != nil {
		return "", err
	}

	targetDir := outputPath + "/" + torrentName
	if _, err := os.ReadDir(targetDir); err != nil {
		log.Printf("creating %s directory to store downloads", targetDir)
		if err := os.Mkdir(targetDir, 0755); err != nil {
			return "", err
		}
	}

	return targetDir, nil
}

// printResultDetails outputs a list of downloaded files with their sizes
func printResultDetails(dir string) error {
	files, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	fmt.Printf("---\nDownload results\n%s:\n", dir)

	for _, file := range files {
		info, err := file.Info()
		if err != nil {
			return err
		}

		mbSize := float64(info.Size()) / (1024 * 1024)
		fmt.Printf("    %s: %.2f MB\n", info.Name(), mbSize)
	}

	return nil
}
