package torrentfile

import (
	"bytes"
	"crypto/rand"
	"crypto/sha1"
	"fmt"
	"io"
	"net"

	"github.com/cedrickchee/torrn/p2p"
	"github.com/jackpal/bencode-go"
)

// Port to listen on
const port = 6881

// Torrent encodes the metadata from a .torrent file
type Torrent struct {
	Name        string
	Announce    string
	InfoHash    [20]byte
	PieceHashes [][20]byte
	PieceLength int
	Length      int
}

type bencodeInfo struct {
	Length      int    `bencode:"length"`
	Name        string `bencode:"name"`
	PieceLength int    `bencode:"piece length"`
	Pieces      string `bencode:"pieces"`
}

type bencodeTorrent struct {
	Announce string      `bencode:"announce"`
	Info     bencodeInfo `bencode:"info"`
}

// Open parses a torrent file.
func Open(r io.Reader) (*Torrent, error) {
	bto := bencodeTorrent{}
	err := bencode.Unmarshal(r, &bto)
	if err != nil {
		return nil, err
	}
	t, err := bto.toTorrent()
	if err != nil {
		return nil, err
	}

	return t, nil
}

// Download downloads a torrent
func (t *Torrent) Download() error {
	// A PeerID is a 20 byte unique identifier presented to trackers and peers
	var peerID [20]byte
	_, err := rand.Read(peerID[:])
	if err != nil {
		return err
	}

	// peers, err := t.getPeers(peerID, port)
	// fmt.Println(peers)
	localhost := p2p.Peer{
		IP:   net.IP{127, 0, 0, 1},
		Port: 51413,
	}
	err = p2p.Connect(&localhost, peerID, t.InfoHash)
	if err != nil {
		return err
	}

	return nil
}

func (i *bencodeInfo) hash() ([20]byte, error) {
	var buf bytes.Buffer
	err := bencode.Marshal(&buf, *i)
	if err != nil {
		return [20]byte{}, err
	}
	hs := sha1.Sum(buf.Bytes())
	return hs, nil
}

func (i *bencodeInfo) splitPieceHashes() ([][20]byte, error) {
	hashLen := 20 // length of SHA-1 hash
	buf := []byte(i.Pieces)
	if len(buf)%hashLen != 0 {
		err := fmt.Errorf("Received malformed pieces of length %d", len(buf))
		return nil, err
	}
	numHashes := len(buf) / hashLen
	hashes := make([][20]byte, numHashes)

	for i := 0; i < numHashes; i++ {
		copy(hashes[i][:], buf[i*hashLen:(i+1)*hashLen])
	}

	return hashes, nil
}

func (bto *bencodeTorrent) toTorrent() (*Torrent, error) {
	infoHash, err := bto.Info.hash()
	if err != nil {
		return nil, err
	}
	pieceHashes, err := bto.Info.splitPieceHashes()
	if err != nil {
		return nil, err
	}

	t := Torrent{
		Name:        bto.Info.Name,
		Announce:    bto.Announce,
		InfoHash:    infoHash,
		PieceHashes: pieceHashes,
		PieceLength: bto.Info.PieceLength,
		Length:      bto.Info.Length,
	}

	return &t, nil
}
