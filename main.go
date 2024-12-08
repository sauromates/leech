package main

import (
	"io"
	"log"
	"os"

	"gihub.com/sauromates/leech/torrentfile"
)

func main() {
	log.SetOutput(io.Discard)

	inputPath := os.Args[1]
	outputPath := os.Args[2]

	torrentfile, err := torrentfile.Open(inputPath)
	if err != nil {
		log.Fatal(err)
	}

	if err := torrentfile.Download(outputPath); err != nil {
		log.Fatal(err)
	}
}
