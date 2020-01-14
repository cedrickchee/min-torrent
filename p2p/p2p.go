package p2p

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"log"
	"net"
	"runtime"

	"github.com/cedrickchee/torrn/message"
)

const maxBlockSize = 32768
const maxUnfulfilled = 5

// Peer encodes connection information for connecting to a peer
type Peer struct {
	IP   net.IP
	Port uint16
}

// Torrent holds data required to download a torrent from a list of peers
type Torrent struct {
	Peers       []Peer
	PeerID      [20]byte
	InfoHash    [20]byte
	PieceHashes [][20]byte
	Length      int
}

type pieceWork struct {
	index  int
	hash   [20]byte
	length int
}

type pieceResult struct {
	index int
	buf   []byte
}

// Download downloads a torrent
func (t *Torrent) Download() ([]byte, error) {
	numPieces := len(t.PieceHashes)

	// Init queues for workers to retrieve work and send results
	workQueue := make(chan *pieceWork, numPieces)
	results := make(chan *pieceResult, numPieces)
	for index, hash := range t.PieceHashes {
		length := t.Length / numPieces
		workQueue <- &pieceWork{index, hash, length}
	}

	// Start workers
	for _, peer := range t.Peers {
		go t.downloadWorker(peer, workQueue, results)
	}

	// Collect results into a buffer until full
	buf := make([]byte, t.Length)
	donePieces := 0
	for donePieces < numPieces {
		res := <-results
		begin, end := calculateBoundsForPiece(res.index, numPieces, t.Length)
		copy(buf[begin:end], res.buf)
		donePieces++

		percent := float64(donePieces) / float64(numPieces)
		log.Printf("(%0.2f%%) Downloaded piece #%d with %d goroutines\n", percent, res.index, runtime.NumGoroutine())
	}
	close(workQueue)

	return buf, nil
}

func (t *Torrent) downloadWorker(peer Peer, workQueue chan *pieceWork, results chan *pieceResult) {
	c, err := newClient(peer, t.PeerID, t.InfoHash)
	if err != nil {
		log.Printf("Peer %s unresponsive. Disconnecting\n", peer.IP)
		return
	}
	defer c.conn.Close()

	c.unchoke()
	c.interested()

	for pw := range workQueue {
		if !c.hasPiece(pw.index) {
			// Re-enqueue the piece to try again
			workQueue <- pw // Put piece back on the queue
			continue
		}

		// Download the piece
		buf, err := attemptDownloadPiece(c, pw)
		if err != nil {
			log.Println("Exiting", err)
			workQueue <- pw // Put piece back on the queue
			return
		}

		err = checkIntegrity(pw, buf)
		if err != nil {
			log.Printf("Piece #%d failed integrity check\n", pw.index)
			workQueue <- pw // Put piece back on the queue
			continue
		}

		results <- &pieceResult{pw.index, buf}
		c.have(pw.index)
	}
}

func attemptDownloadPiece(c *client, pw *pieceWork) ([]byte, error) {
	pieceLength := pw.length
	buf := make([]byte, pieceLength)
	downloaded := 0
	requested := 0
	for downloaded < pieceLength {
		for c.hasNext() {
			msg, err := c.read() // this call blocks
			if err != nil {
				return nil, err
			}
			if msg == nil { // keep-alive
				continue
			}
			if msg.ID != message.MsgPiece {
				log.Println(msg)
			}
			switch msg.ID {
			case message.MsgUnchoke:
				c.choked = false
			case message.MsgChoke:
				c.choked = true
			case message.MsgHave:
				index, err := message.ParseHave(msg)
				if err != nil {
					return nil, err
				}
				c.bitfield.SetPiece(index)
			case message.MsgPiece:
				n, err := message.ParsePiece(pw.index, buf, msg)
				if err != nil {
					return nil, err
				}
				downloaded += n
			}
		}

		if !c.choked && requested < pieceLength && requested-downloaded <= maxUnfulfilled+1 {
			for i := 0; i < maxUnfulfilled; i++ {
				blockSize := maxBlockSize
				if pieceLength-requested < blockSize {
					// Last block might be shorter than the typical block
					blockSize = pieceLength - requested
				}
				// fmt.Println("Request")
				c.request(pw.index, requested, blockSize)
				requested += blockSize
			}
		}

		msg, err := c.read() // this call blocks
		if err != nil {
			return nil, err
		}
		if msg == nil { // keep-alive
			continue
		}
		if msg.ID != message.MsgPiece {
			log.Println(msg)
		}

		switch msg.ID {
		case message.MsgChoke:
			c.choked = true
		case message.MsgUnchoke:
			c.choked = false
		case message.MsgHave:
			index, err := message.ParseHave(msg)
			if err != nil {
				return nil, err
			}
			c.bitfield.SetPiece(index)
		case message.MsgPiece:
			n, err := message.ParsePiece(pw.index, buf, msg)
			if err != nil {
				return nil, err
			}
			downloaded += n
		}
	}
	return buf, nil
}

func calculateBoundsForPiece(index, numPieces, length int) (begin int, end int) {
	pieceLength := length / numPieces
	begin = index * pieceLength
	end = begin + pieceLength
	return begin, end
}

func checkIntegrity(pw *pieceWork, buf []byte) error {
	hash := sha1.Sum(buf)

	if !bytes.Equal(hash[:], pw.hash[:]) {
		return fmt.Errorf("Index %d failed integrity check", pw.index)
	}
	return nil
}
