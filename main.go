package main

import (
	"fmt"
	"log"
	"os"

	"github.com/sauromates/leech/torrent"
	"github.com/sauromates/leech/torrentfile"
)

func main() {
	if err := configureLogs("leech.log"); err != nil {
		log.Fatal(err)
	}

	inputPath := os.Args[1]
	torrentfile, err := torrentfile.Open(inputPath)
	if err != nil {
		log.Fatal(err)
	}

	dir, err := createDownloadDir(torrentfile.Name)
	if err != nil {
		log.Fatal(err)
	}

	torrent, err := torrent.CreateFromTorrentFile(torrentfile)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Downloading\n---\n%s\n", torrent)

	if err := torrent.Download(dir); err != nil {
		log.Fatal(err)
	}

	printResultDetails(dir)
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
