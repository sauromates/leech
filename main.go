package main

import (
	"fmt"
	"log"
	"os"
	"text/template"

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

	var downloadError error

	if len(torrentfile.Paths) > 0 {
		if err := printTorrentDetails(torrentfile); err != nil {
			log.Fatal(err)
		}

		dir, err := createDownloadDir(torrentfile.Name)
		if err != nil {
			log.Fatal(err)
		}

		downloadError = torrentfile.DownloadMultiple(dir)
	} else {
		downloadError = torrentfile.Download(torrentfile.Name)
	}

	if downloadError != nil {
		log.Fatal(downloadError)
	}
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

// printTorrentDetails parses torrent file and prints detailed information
// about it using default `tmpl` file
func printTorrentDetails(tf torrentfile.TorrentFile) error {
	decoded, err := tf.Parse()
	if err != nil {
		return err
	}

	funcs := template.FuncMap{
		"toMB": func(size int) string {
			return fmt.Sprintf("%.2f MB", float64(size/(1024*1024)))
		},
	}
	template, err := template.New("torrent.go.tmpl").Funcs(funcs).ParseFiles("templates/torrent.go.tmpl")
	if err != nil {
		return err
	}

	return template.Execute(os.Stdout, decoded)
}
