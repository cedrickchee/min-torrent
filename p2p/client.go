package p2p

import (
	"fmt"
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/cedrickchee/torrn/handshake"
	"github.com/cedrickchee/torrn/message"
)

type client struct {
	conn     net.Conn
	bitfield message.Bitfield
	Choked   bool
	Mux      sync.Mutex
}

func completeHandshake(conn net.Conn, infoHash, peerID [20]byte) (*handshake.Handshake, error) {
	conn.SetDeadline(time.Now().Local().Add(3 * time.Second))
	defer conn.SetDeadline(time.Time{}) // disable the deadline

	req := handshake.New(infoHash, peerID)
	_, err := conn.Write(req.Serialize())
	if err != nil {
		return nil, err
	}

	res, err := handshake.Read(conn)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func recvBitfield(conn net.Conn) (message.Bitfield, error) {
	conn.SetDeadline(time.Now().Local().Add(3 * time.Second))
	defer conn.SetDeadline(time.Time{}) // disable the deadline

	msg, err := message.Read(conn)
	if err != nil {
		return nil, err
	}
	if msg.ID != message.MsgBitfield {
		err := fmt.Errorf("Expected bitfield but got ID %d", msg.ID)
		return nil, err
	}

	return msg.Payload, nil
}

func newClient(peer Peer, peerID, infoHash [20]byte) (*client, error) {
	// Connect
	hostPort := net.JoinHostPort(peer.IP.String(), strconv.Itoa(int(peer.Port)))
	conn, err := net.DialTimeout("tcp", hostPort, 3*time.Second)
	if err != nil {
		return nil, err
	}
	fmt.Println("Connected to:", hostPort)

	// Handshake
	h, err := completeHandshake(conn, infoHash, peerID)
	if err != nil {
		return nil, err
	}
	fmt.Println("Handshake:", h)

	// Get bitfield
	bf, err := recvBitfield(conn)
	if err != nil {
		return nil, err
	}

	return &client{
		conn:     conn,
		bitfield: bf,
		Choked:   true,
		Mux:      sync.Mutex{},
	}, nil
}

func (c *client) hasPiece(index int) bool {
	return c.bitfield.HasPiece(index)
}
