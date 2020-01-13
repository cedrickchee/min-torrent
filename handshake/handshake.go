package handshake

import (
	"errors"
	"io"

	"github.com/cedrickchee/torrn/torrentfile"
)

// A Handshake is a sequence of bytes a peer uses to identify itself
type Handshake struct {
	Pstr     string             // the protocol identifier
	InfoHash [20]byte           // which file we want
	PeerID   torrentfile.PeerID // made up ID to identify ourselves
}

// Serialize serializes the handshake to a buffer
//
// BitTorrent handshake is made up of five parts:
// 1. The length of the protocol identifier, which is always 19 (0x13 in hex)
// 2. The protocol identifier, called the pstr which
// is always 'BitTorrent protocol'
// 3. Eight reserved bytes, all set to 0. We’d flip some of them to 1
// to indicate that we support certain extensions.
// 4. The infohash that we calculated earlier to identify which file we want
// 5. The Peer ID that we made up to identify ourselves
//
// Put together, a handshake string might look like this:
// \x13BitTorrent protocol\x00\x00\x00\x00\x00\x00\x00\x00\x86\xd4\xc8\x00\x24\xa4\x69\xbe\x4c\x50\xbc\x5a\x10\x2c\xf7\x17\x80\x31\x00\x74-TR2940-k8hj0wgej6ch
func (h *Handshake) Serialize() []byte {
	pstrLen := len(h.Pstr)
	bufLen := 49 + pstrLen
	buf := make([]byte, bufLen)
	buf[0] = byte(pstrLen)
	copy(buf[1:], h.Pstr)
	// Leave 8 reserved bytes
	copy(buf[1+pstrLen+8:], h.InfoHash[:])
	copy(buf[1+pstrLen+8+20:], h.PeerID[:])

	return buf
}

// ReadHandshake parses a message from a stream. Returns `nil` on keep-alive message
func ReadHandshake(r io.Reader) (*Handshake, error) {
	// Do Serialize(), but backwards
	lengthBuf := make([]byte, 1)
	_, err := io.ReadFull(r, lengthBuf)
	if err != nil {
		return nil, err
	}
	pstrLen := int(lengthBuf[0])

	if pstrLen == 0 {
		err := errors.New("pstrLen cannot be 0")
		return nil, err
	}

	handshakeBuf := make([]byte, 48+pstrLen)
	_, err = io.ReadFull(r, handshakeBuf)
	if err != nil {
		return nil, err
	}

	var infoHash [20]byte
	var peerID torrentfile.PeerID

	copy(infoHash[:], handshakeBuf[pstrLen+8:pstrLen+8+20])
	copy(peerID[:], handshakeBuf[pstrLen+8+20:])

	h := Handshake{
		Pstr:     string(handshakeBuf[0:pstrLen]),
		InfoHash: infoHash,
		PeerID:   peerID,
	}

	return &h, nil
}
