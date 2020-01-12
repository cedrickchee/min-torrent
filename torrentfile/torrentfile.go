package torrentfile

import "io"

import "github.com/jackpal/bencode-go"

const port = 6881

type Info struct {
	Length      int    `bencode:"length"`
	Name        string `bencode:"name"`
	PieceLength int    `bencode:"piece length"`
	Pieces      string `bencode:"pieces"`
}

type Torrent struct {
	Announce string `bencode:"announce"`
	Info     Info   `bencode:"info"`
}

// Open parses a torrent file.
func Open(r io.Reader) (*Torrent, error) {
	to := Torrent{}
	err := bencode.Unmarshal(r, &to)
	if err != nil {
		return nil, err
	}
	return &to, nil
}
