package client

import (
	"bufio"
	"bytes"
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/cedrickchee/torrn/bitfield"
	"github.com/cedrickchee/torrn/handshake"
	"github.com/cedrickchee/torrn/message"
	"github.com/cedrickchee/torrn/peers"
)

type Client struct {
	Conn     net.Conn
	reader   *bufio.Reader
	Bitfield bitfield.Bitfield
	Choked   bool
}

func completeHandshake(conn net.Conn, r *bufio.Reader, infoHash, peerID [20]byte) (*handshake.Handshake, error) {
	conn.SetDeadline(time.Now().Add(3 * time.Second))
	defer conn.SetDeadline(time.Time{}) // disable the deadline

	req := handshake.New(infoHash, peerID)
	_, err := conn.Write(req.Serialize())
	if err != nil {
		return nil, err
	}

	res, err := handshake.Read(r)
	if err != nil {
		return nil, err
	}
	if !bytes.Equal(res.InfoHash[:], infoHash[:]) {
		return nil, fmt.Errorf("Expected infohash %x but got %x", res.InfoHash, infoHash)
	}
	return res, nil
}

func recvBitfield(conn net.Conn, r *bufio.Reader) (bitfield.Bitfield, error) {
	conn.SetDeadline(time.Now().Add(3 * time.Second))
	defer conn.SetDeadline(time.Time{}) // disable the deadline

	msg, err := message.Read(r)
	if err != nil {
		return nil, err
	}
	if msg.ID != message.MsgBitfield {
		err := fmt.Errorf("Expected bitfield but got ID %d", msg.ID)
		return nil, err
	}

	return msg.Payload, nil
}

// New connects with a peer, completes a handshake, and receives a handshake
func New(peer peers.Peer, peerID, infoHash [20]byte) (*Client, error) {
	// Connect
	hostPort := net.JoinHostPort(peer.IP.String(), strconv.Itoa(int(peer.Port)))
	conn, err := net.DialTimeout("tcp", hostPort, 3*time.Second)
	if err != nil {
		return nil, err
	}

	reader := bufio.NewReader(conn)

	// Handshake
	_, err = completeHandshake(conn, reader, infoHash, peerID)
	if err != nil {
		conn.Close()
		return nil, err
	}

	// Get bitfield
	bf, err := recvBitfield(conn, reader)
	if err != nil {
		conn.Close()
		return nil, err
	}

	return &Client{
		Conn:     conn,
		reader:   reader,
		Bitfield: bf,
		Choked:   true,
	}, nil
}

// HasNext returns true if there are unread messages from the peer
func (c *Client) HasNext() bool {
	return c.reader.Buffered() > 0
}

// Read reads and consumes a message from the connection
func (c *Client) Read() (*message.Message, error) {
	msg, err := message.Read(c.reader)
	return msg, err
}

// SendRequest sends a Request message to the peer
func (c *Client) SendRequest(index, begin, length int) error {
	req := message.FormatRequest(index, begin, length)
	_, err := c.Conn.Write(req.Serialize())
	return err
}

// SendHave sends a Have message to the peer
func (c *Client) SendHave(index int) error {
	msg := message.FormatHave(index)
	_, err := c.Conn.Write(msg.Serialize())
	return err
}

// SendInterested sends an Interested message to the peer
func (c *Client) SendInterested() error {
	msg := message.Message{ID: message.MsgInterested}
	_, err := c.Conn.Write(msg.Serialize())
	return err
}

// SendNotInterested sends a NotInterested message to the peer
func (c *Client) SendNotInterested() error {
	msg := message.Message{ID: message.MsgNotInterested}
	_, err := c.Conn.Write(msg.Serialize())
	return err
}

// SendUnchoke sends an Unchoke message to the peer
func (c *Client) SendUnchoke() error {
	msg := message.Message{ID: message.MsgUnchoke}
	_, err := c.Conn.Write(msg.Serialize())
	return err
}
