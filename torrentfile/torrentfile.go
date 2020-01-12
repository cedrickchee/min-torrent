package torrentfile

import (
	"bytes"
	"crypto/rand"
	"crypto/sha1"
	"fmt"
	"io"

	"github.com/jackpal/bencode-go"
)

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

// Download downloads a torrent
func (to *Torrent) Download() error {
	peerID := make([]byte, 20)
	_, err := rand.Read(peerID)
	if err != nil {
		return err
	}

	tracker := Tracker{
		PeerID:  peerID,
		Torrent: to,
		Port:    port,
	}
	peers, err := tracker.getPeers()
	fmt.Println(peers)
	return nil
}

func (i *Info) hash() ([]byte, error) {
	var buf bytes.Buffer
	err := bencode.Marshal(&buf, *i)
	if err != nil {
		return nil, err
	}
	hs := sha1.Sum(buf.Bytes())
	return hs[:], nil
}
