package main

import (
	"log"
	"os"

	"github.com/cedrickchee/torrn/torrentfile"
)

func main() {
	inPath := os.Args[1]
	outPath := os.Args[2]

	inFile, err := os.Open(inPath)
	checkError(err)
	defer inFile.Close()

	t, err := torrentfile.Open(inFile)
	checkError(err)
	buf, err := t.Download()
	checkError(err)

	outFile, err := os.Create(outPath)
	checkError(err)
	defer outFile.Close()
	_, err = outFile.Write(buf)
	checkError(err)
}

func checkError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
