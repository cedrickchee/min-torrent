package torrentfile

import (
	"encoding/binary"
	"errors"
	"net"
	"net/http"
	"net/url"
	"strconv"

	"github.com/jackpal/bencode-go"
)

// Peer encodes information for connecting to a peer
type Peer struct {
	IP   net.IP
	Port uint16
}

// Tracker
// type Tracker struct {
// 	PeerID  []byte
// 	Torrent *Torrent
// 	Port    uint16
// }

// TrackerResponse
type bencodeTrackerResponse struct {
	Interval int    `bencode:"interval"`
	Peers    string `bencode:"port"`
}

func (t *Torrent) getPeers(peerID []byte, port uint16) ([]Peer, error) {
	url, err := t.buildTrackerURL(peerID, port)
	if err != nil {
		return nil, err
	}

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	trackerResp := bencodeTrackerResponse{}
	err = bencode.Unmarshal(resp.Body, &trackerResp)
	if err != nil {
		return nil, err
	}

	peers, err := parsePeers(trackerResp.Peers)
	if err != nil {
		return nil, err
	}

	return peers, nil
}

func (t *Torrent) buildTrackerURL(peerID []byte, port uint16) (string, error) {
	base, err := url.Parse(t.Announce)
	if err != nil {
		return "", err
	}

	params := url.Values{
		"info_hash":  []string{string(t.InfoHash)},
		"peer_id":    []string{string(peerID)},
		"port":       []string{strconv.Itoa(int(port))},
		"uploaded":   []string{"0"},
		"downloaded": []string{"0"},
		"compact":    []string{"1"},
		"left":       []string{strconv.Itoa(t.Length)},
	}
	base.RawQuery = params.Encode()

	return base.String(), nil
}

func parsePeers(peersBin string) ([]Peer, error) {
	peerSize := 6 // 4 for IP, 2 for port
	numPeers := len(peersBin) / peerSize
	if len(peersBin)%peerSize != 0 {
		err := errors.New("Received malformed peers")
		return nil, err
	}
	peers := make([]Peer, numPeers)
	for i := 0; i < numPeers; i++ {
		offset := i * peerSize
		peers[i].IP = net.IP(peersBin[offset : offset+4])
		peers[i].Port = binary.BigEndian.Uint16([]byte(peersBin[offset+4 : offset+6]))
	}

	return peers, nil
}
