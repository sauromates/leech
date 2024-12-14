package main

import (
	"fmt"
	"log"
	"os"
	"text/template"

	"gihub.com/sauromates/leech/torrentfile"
)

func main() {
	inputPath := os.Args[1]
	torrentfile, err := torrentfile.Open(inputPath)
	if err != nil {
		log.Fatal(err)
	}

	if err := printTorrentDetails(torrentfile); err != nil {
		log.Fatal(err)
	}

	var downloadError error
	if len(torrentfile.Paths) > 0 {
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

func printTorrentDetails(torrent torrentfile.TorrentFile) error {
	decoded, err := torrent.Parse()
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
