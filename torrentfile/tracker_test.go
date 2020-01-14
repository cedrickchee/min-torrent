package torrentfile

import (
	"net"
	"testing"

	"github.com/cedrickchee/torrn/p2p"
	"github.com/stretchr/testify/assert"
)

func TestParsePeers(t *testing.T) {
	tests := map[string]struct {
		input  string
		output []p2p.Peer
		fails  bool
	}{
		"correctly parse peers": {
			input: string([]byte{127, 0, 0, 1, 0x00, 0x50, 1, 1, 1, 1, 0x01, 0xbb}),
			output: []p2p.Peer{
				{IP: net.IP{127, 0, 0, 1}, Port: 80},
				{IP: net.IP{1, 1, 1, 1}, Port: 443},
			},
			fails: false,
		},
		"not enough bytes in peers": {
			input:  string([]byte{127, 0, 0, 1, 0x00}),
			output: nil,
			fails:  true,
		},
	}

	for _, test := range tests {
		peers, err := parsePeers(test.input)
		if test.fails {
			assert.NotNil(t, err)
		} else {
			assert.Nil(t, err)
		}
		assert.Equal(t, test.output, peers)
	}
}

func TestBuildTrackerURL(t *testing.T) {
	to := TorrentFile{
		Name:     "debian-10.2.0-amd64-netinst.iso",
		Announce: "http://bttracker.debian.org:6969/announce",
		InfoHash: [20]byte{216, 247, 57, 206, 195, 40, 149, 108, 204, 91, 191, 31, 134, 217, 253, 207, 219, 168, 206, 182},
		PieceHashes: [][20]byte{
			{49, 50, 51, 52, 53, 54, 55, 56, 57, 48, 97, 98, 99, 100, 101, 102, 103, 104, 105, 106},
			{97, 98, 99, 100, 101, 102, 103, 104, 105, 106, 49, 50, 51, 52, 53, 54, 55, 56, 57, 48},
		},
		PieceLength: 262144,
		Length:      351272960,
	}
	peerID := [20]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20}
	var port uint16 = 6882
	url, err := to.buildTrackerURL(peerID, port)
	expected := "http://bttracker.debian.org:6969/announce?compact=1&downloaded=0&info_hash=%D8%F79%CE%C3%28%95l%CC%5B%BF%1F%86%D9%FD%CF%DB%A8%CE%B6&left=351272960&peer_id=%01%02%03%04%05%06%07%08%09%0A%0B%0C%0D%0E%0F%10%11%12%13%14&port=6882&uploaded=0"
	assert.Nil(t, err)
	assert.Equal(t, expected, url)
}
