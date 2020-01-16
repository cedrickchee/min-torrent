package main

import (
	"log"
	"os"

	"github.com/cedrickchee/min-torrent/torrentfile"
)

func main() {
	inPath := os.Args[1]
	outPath := os.Args[2]

	t, err := torrentfile.Open(inPath)
	checkError(err)
	err = t.DownloadToFile(outPath)
	checkError(err)
}

func checkError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
