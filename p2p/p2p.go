package p2p

import (
	"crypto/sha1"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math"
	"net"
	"strconv"

	"github.com/cedrickchee/torrn/handshake"
	"github.com/cedrickchee/torrn/message"
)

// Peer encodes connection information for connecting to a peer
type Peer struct {
	IP   net.IP
	Port uint16
}

// Downloader holds data required to download a torrent from a list of peers
type Downloader struct {
	Peers       []Peer
	PeerID      [20]byte
	InfoHash    [20]byte
	PieceHashes [][20]byte
	Length      int
}

func (d *Downloader) Download() error {
	conn, err := d.Peers[0].connect(d.PeerID, d.InfoHash)
	defer conn.Close()
	if err != nil {
		fmt.Println("Error", err)
		return err
	}
	h, err := d.handshake(conn)
	if err != nil {
		return err
	}
	fmt.Println("Handshake:", h)

	choked := false
	pieceSize := d.Length / len(d.PieceHashes)
	buf := make([]byte, pieceSize)
	i := 0
	for i < pieceSize {
		msg, err := message.Read(conn)
		if err != nil {
			return err
		}

		switch msg.ID {
		case message.MsgChoke:
			choked = true
		case message.MsgUnchoke:
			choked = false
		case message.MsgPiece:
			begin := binary.BigEndian.Uint32(msg.Payload[4:8])
			copy(buf[begin:], msg.Payload[8:])
			i += (len(msg.Payload) - 8)
		}

		if !choked {
			fmt.Printf("Downloading: %v of %v\n", i, pieceSize)
			index := 0 // Piece number
			begin := i // Offset
			remain := pieceSize - i
			length := int(math.Min(float64(16384), float64(pieceSize)))
			length = int(math.Min(float64(remain), float64(length)))
			_, err := conn.Write(message.FormatRequest(index, begin, length).Serialize())
			if err != nil {
				return err
			}
		}
	}

	s := sha1.Sum(buf)
	fmt.Println("Buffer hash:", hex.EncodeToString(s[:]))
	fmt.Println("Torrent file piece hash:", hex.EncodeToString(d.PieceHashes[0][:]))

	return nil
}

func (p *Peer) connect(peerID [20]byte, infoHash [20]byte) (net.Conn, error) {
	hostPort := net.JoinHostPort(p.IP.String(), strconv.Itoa(int(p.Port)))
	conn, err := net.Dial("tcp", hostPort)
	if err != nil {
		return nil, err
	}
	fmt.Println("Connected to:", hostPort)
	return conn, nil
}

func (d *Downloader) handshake(conn net.Conn) (*handshake.Handshake, error) {
	req := handshake.Handshake{
		Pstr:     "BitTorrent protocol",
		InfoHash: d.InfoHash,
		PeerID:   d.PeerID,
	}
	_, err := conn.Write(req.Serialize())
	if err != nil {
		return nil, err
	}

	res, err := handshake.ReadHandshake(conn)
	if err != nil {
		return nil, err
	}
	return res, nil
}
