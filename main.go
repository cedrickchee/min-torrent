package main

import (
	"log"
	"os"

	"github.com/cedrickchee/torrn/torrentfile"
)

func main() {
	file, err := os.Open(os.Args[1])
	checkError(err)
	defer file.Close()

	t, err := torrentfile.Open(file)
	checkError(err)
	err = t.Download()
	checkError(err)
}

func checkError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
