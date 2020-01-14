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

/*type swarm struct {
	clients    []*client
	queue      chan *pieceWork
	buf        []byte
	piecesDone int
	mux        sync.Mutex
}*/

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

	// clients := d.initClients()
	// if len(clients) == 0 {
	// 	return nil, fmt.Errorf("Could not connect to any of %d clients", len(d.Peers))
	// }
	// log.Printf("Connected to %d clients\n", len(clients))

	// queue := make(chan *pieceWork, len(t.PieceHashes))
	// for index, hash := range t.PieceHashes {
	// 	queue <- &pieceWork{index, hash}
	// }
	// return t.processQueue(clients, queue), nil
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
	// Previously known as downloadPiece
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
		} else {
			// log.Println("Received", len(msg.Payload), "bytes")
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

/*func (t *Torrent) initClients() []*client {
	// Create clients in parallel
	c := make(chan *client)
	for _, p := range t.Peers {
		go func(p Peer) {
			client, err := newClient(p, t.PeerID, t.InfoHash)
			if err != nil {
				c <- nil
			} else {
				c <- client
			}
		}(p)
	}

	// Gather clients into a slice
	clients := make([]*client, 0)
	for range t.Peers {
		client := <-c
		if client != nil {
			clients = append(clients, client)
		}
	}

	return clients
}

func (t *Torrent) processQueue(clients []*client, queue chan *pieceWork) []byte {
	s := swarm{
		clients: clients,
		queue:   queue,
		buf:     make([]byte, t.Length),
		mux:     sync.Mutex{},
	}

	numWorkers := (len(s.clients) + 1) / 2
	log.Printf("Spawning %d workers\n", numWorkers)
	wg := sync.WaitGroup{}
	wg.Add(numWorkers)

	for i := 0; i < numWorkers; i++ {
		go s.worker(t, &wg)
	}

	wg.Wait()
	return s.buf
}

func (s *swarm) selectClient(index int) (*client, error) {
	for _, c := range s.clients {
		if !c.engaged && !c.choked && c.hasPiece(index) {
			return c, nil
		}
	}
	for _, c := range s.clients {
		if !c.engaged && c.hasPiece(index) {
			return c, nil
		}
	}
	return nil, fmt.Errorf("Could not find client for piece %d", index)
}

func (s *swarm) worker(t *Torrent, wg *sync.WaitGroup) {
	for pw := range s.queue {
		s.mux.Lock()
		c, err := s.selectClient(pw.index)
		if err != nil {
			// fmt.Println("Error:", err)
			// Re-enqueue the piece to try again
			s.queue <- pw
			s.mux.Unlock()
			continue
		}
		c.engaged = true
		s.mux.Unlock()

		begin, end := calculateBoundsForPiece(pw.index, len(t.PieceHashes), t.Length)
		pieceLength := end - begin
		pieceBuf, err := downloadPiece(c, pw, pieceLength)
		if err != nil {
			// Re-enqueue the piece to try again
			log.Println(err)
			s.removeClient(c)
			s.queue <- pw
		} else {
			// Copy into buffer should not overlap with other workers
			copy(s.buf[begin:end], pieceBuf)

			s.mux.Lock()
			s.piecesDone++
			log.Printf("Downloaded piece %d (%d/%d) %0.2f%%\n", pw.index, s.piecesDone, len(t.PieceHashes), float64(s.piecesDone)/float64(len(t.PieceHashes))*100)
			s.mux.Unlock()

			if s.piecesDone == len(t.PieceHashes) {
				close(s.queue)
			}
		}

		s.mux.Lock()
		c.engaged = false
		s.mux.Unlock()
	}
	wg.Done()
}*/

func calculateBoundsForPiece(index, numPieces, length int) (begin int, end int) {
	pieceLength := length / numPieces
	begin = index * pieceLength
	end = begin + pieceLength
	return begin, end
}

/*func downloadPiece(c *client, pw *pieceWork, pieceLength int) ([]byte, error) {
	buf := make([]byte, pieceLength)
	c.unchoke()
	c.interested()
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
				fmt.Println(msg)
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

		if !c.choked && requested < pieceLength && requested-downloaded <= queueSize+1 {
			for i := 0; i < queueSize; i++ {
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
			fmt.Println(msg)
		} else {
			fmt.Println("Received", len(msg.Payload), "bytes")
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

	c.have(pw.index)
	c.notInterested()

	err := checkIntegrity(pw, buf)
	if err != nil {
		return nil, err
	}

	return buf, nil
}*/

func checkIntegrity(pw *pieceWork, buf []byte) error {
	hash := sha1.Sum(buf)

	if !bytes.Equal(hash[:], pw.hash[:]) {
		return fmt.Errorf("Index %d failed integrity check", pw.index)
	}
	return nil
}

/*func (s *swarm) removeClient(c *client) {
	if len(s.clients) == 1 {
		panic("Removed last client")
	}
	log.Printf("Removing client. %d clients remaining\n", len(s.clients))
	s.mux.Lock()
	c.conn.Close()
	var i int
	for i = 0; i < len(s.clients); i++ {
		if s.clients[i] == c {
			break
		}
	}
	s.clients = append(s.clients[:i], s.clients[i+1:]...)
	s.mux.Unlock()
}*/
