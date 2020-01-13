package torrentfile

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestToTorrent(t *testing.T) {
	tests := map[string]struct {
		input  *bencodeTorrent
		output *Torrent
		fails  bool
	}{
		"correct conversion": {
			input: &bencodeTorrent{
				Announce: "http://bttracker.debian.org:6969/announce",
				Info: bencodeInfo{
					Length:      351272960,
					Name:        "debian-10.2.0-amd64-netinst.iso",
					PieceLength: 262144,
					Pieces:      "1234567890abcdefghijabcdefghij1234567890",
				},
			},
			output: &Torrent{
				Name:     "debian-10.2.0-amd64-netinst.iso",
				Announce: "http://bttracker.debian.org:6969/announce",
				InfoHash: [20]byte{216, 247, 57, 206, 195, 40, 149, 108, 204, 91, 191, 31, 134, 217, 253, 207, 219, 168, 206, 182},
				PieceHashes: [][20]byte{
					{49, 50, 51, 52, 53, 54, 55, 56, 57, 48, 97, 98, 99, 100, 101, 102, 103, 104, 105, 106},
					{97, 98, 99, 100, 101, 102, 103, 104, 105, 106, 49, 50, 51, 52, 53, 54, 55, 56, 57, 48},
				},
				PieceLength: 262144,
				Length:      351272960,
			},
			fails: false,
		},
		"not enough bytes in pieces": {
			input: &bencodeTorrent{
				Announce: "http://bttracker.debian.org:6969/announce",
				Info: bencodeInfo{
					Length:      351272960,
					Name:        "debian-10.2.0-amd64-netinst.iso",
					PieceLength: 262144,
					Pieces:      "1234567890abcdefghijabcdef", // Only 26 bytes
				},
			},
			output: nil,
			fails:  true,
		},
	}

	for _, test := range tests {
		to, err := test.input.toTorrent()
		if test.fails {
			assert.NotNil(t, err)
		} else {
			assert.Nil(t, err)
		}
		assert.Equal(t, test.output, to)
	}
}
