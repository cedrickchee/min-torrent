package p2p

import (
	"net"
	"strconv"
)

// Peer encodes connection information for connecting to a peer
type Peer struct {
	IP   net.IP
	Port uint16
}

// Downloader holds data required to download a torrent from a list of peers
type Downloader struct {
	Peers       []Peer
	InfoHash    [20]byte
	PieceLength int
	Length      int
}

func (d *Downloader) Download() error {
	return nil
}

func connect(p *Peer, peerID [20]byte, infoHash [20]byte) (net.Conn, error) {
	hostPort := net.JoinHostPort(p.IP.String(), strconv.Itoa(int(p.Port)))
	conn, err := net.Dial("tcp", hostPort)
	if err != nil {
		return nil, err
	}
	return conn, nil
}
