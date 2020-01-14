package message

import (
	"bufio"
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSerialize(t *testing.T) {
	tests := map[string]struct {
		input  *Message
		output []byte
	}{
		"serialize message": {
			input:  &Message{ID: MsgHave, Payload: []byte{1, 2, 3, 4}},
			output: []byte{0, 0, 0, 5, 4, 1, 2, 3, 4},
		},
		"serialize keep-alive": {
			input:  nil,
			output: []byte{0, 0, 0, 0},
		},
	}

	for _, test := range tests {
		buf := test.input.Serialize()
		assert.Equal(t, test.output, buf)
	}
}

func TestRead(t *testing.T) {
	tests := map[string]struct {
		input  []byte
		output *Message
		fails  bool
	}{
		"parse normal message into struct": {
			input:  []byte{0, 0, 0, 5, 4, 1, 2, 3, 4},
			output: &Message{ID: MsgHave, Payload: []byte{1, 2, 3, 4}},
			fails:  false,
		},
		"parse keep-alive into nil": {
			input:  []byte{0, 0, 0, 0},
			output: nil,
			fails:  false,
		},
		"length too short": {
			input:  []byte{1, 2, 3},
			output: nil,
			fails:  true,
		},
		"buffer too short for length": {
			input:  []byte{0, 0, 0, 5, 4, 1, 2, 3},
			output: nil,
			fails:  true,
		},
	}

	for _, test := range tests {
		reader := bufio.NewReader(bytes.NewReader(test.input))
		msg, err := Read(reader)
		if test.fails {
			assert.NotNil(t, err)
		} else {
			assert.Nil(t, err)
		}
		assert.Equal(t, test.output, msg)
	}
}

func TestString(t *testing.T) {
	tests := []struct {
		input  *Message
		output string
	}{
		{nil, "KeepAlive"},
		{&Message{MsgChoke, []byte{1, 2, 3}}, "Choke"},
		{&Message{MsgUnchoke, []byte{1, 2, 3}}, "Unchoke"},
		{&Message{MsgInterested, []byte{1, 2, 3}}, "Interested"},
		{&Message{MsgNotInterested, []byte{1, 2, 3}}, "NotInterested"},
		{&Message{MsgHave, []byte{1, 2, 3}}, "Have"},
		{&Message{MsgBitfield, []byte{1, 2, 3}}, "Bitfield"},
		{&Message{MsgRequest, []byte{1, 2, 3}}, "Request"},
		{&Message{MsgPiece, []byte{1, 2, 3}}, "Piece"},
		{&Message{MsgCancel, []byte{1, 2, 3}}, "Cancel"},
		{&Message{MsgPort, []byte{1, 2, 3}}, "Port"},
		{&Message{99, []byte{1, 2, 3}}, "Unknown#99"},
	}

	for _, test := range tests {
		s := test.input.String()
		assert.Equal(t, test.output, s)
	}
}

func TestFormatRequest(t *testing.T) {
	msg := FormatRequest(4, 567, 4321)
	expected := &Message{
		ID: MsgRequest,
		Payload: []byte{
			0x00, 0x00, 0x00, 0x04, // Index
			0x00, 0x00, 0x02, 0x37, // Begin
			0x00, 0x00, 0x10, 0xe1, // Length
		},
	}
	assert.Equal(t, expected, msg)
}

func TestFormatHave(t *testing.T) {
	msg := FormatHave(4)
	expected := &Message{
		ID:      MsgHave,
		Payload: []byte{0x00, 0x00, 0x00, 0x04},
	}
	assert.Equal(t, expected, msg)
}

func TestParsePiece(t *testing.T) {
	tests := map[string]struct {
		inputIndex int
		inputBuf   []byte
		inputMsg   *Message
		outputN    int
		outputBuf  []byte
		fails      bool
	}{
		"parse valid piece": {
			inputIndex: 4,
			inputBuf:   make([]byte, 10),
			inputMsg: &Message{
				ID: MsgPiece,
				Payload: []byte{
					0x00, 0x00, 0x00, 0x04, // Index
					0x00, 0x00, 0x00, 0x02, // Begin
					0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff, // Block
				},
			},
			outputBuf: []byte{0x00, 0x00, 0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff, 0x00, 0x00},
			outputN:   6,
			fails:     false,
		},
		"wrong message type": {
			inputIndex: 4,
			inputBuf:   make([]byte, 10),
			inputMsg: &Message{
				ID:      MsgChoke,
				Payload: []byte{},
			},
			outputBuf: make([]byte, 10),
			outputN:   0,
			fails:     true,
		},
		"payload too short": {
			inputIndex: 4,
			inputBuf:   make([]byte, 10),
			inputMsg: &Message{
				ID: MsgPiece,
				Payload: []byte{
					0x00, 0x00, 0x00, 0x04, // Index
					0x00, 0x00, 0x00, // Malformed offset
				},
			},
			outputBuf: make([]byte, 10),
			outputN:   0,
			fails:     true,
		},
		"wrong index": {
			inputIndex: 4,
			inputBuf:   make([]byte, 10),
			inputMsg: &Message{
				ID: MsgPiece,
				Payload: []byte{
					0x00, 0x00, 0x00, 0x06, // Index is 6, not 4
					0x00, 0x00, 0x00, 0x02, // Begin
					0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff, // Block
				},
			},
			outputBuf: make([]byte, 10),
			outputN:   0,
			fails:     true,
		},
		"offset too high": {
			inputIndex: 4,
			inputBuf:   make([]byte, 10),
			inputMsg: &Message{
				ID: MsgPiece,
				Payload: []byte{
					0x00, 0x00, 0x00, 0x04, // Index is 6, not 4
					0x00, 0x00, 0x00, 0x0c, // Begin is 12 > 10
					0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff, // Block
				},
			},
			outputBuf: make([]byte, 10),
			outputN:   0,
			fails:     true,
		},
		"offset ok but payload too long": {
			inputIndex: 4,
			inputBuf:   make([]byte, 10),
			inputMsg: &Message{
				ID: MsgPiece,
				Payload: []byte{
					0x00, 0x00, 0x00, 0x04, // Index is 6, not 4
					0x00, 0x00, 0x00, 0x02, // Begin is ok
					// Block is 10 long but begin=2; too long for input buffer
					0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff, 0x0a, 0x0b, 0x0c, 0x0d,
				},
			},
			outputBuf: make([]byte, 10),
			outputN:   0,
			fails:     true,
		},
	}

	for _, test := range tests {
		n, err := ParsePiece(test.inputIndex, test.inputBuf, test.inputMsg)
		if test.fails {
			assert.NotNil(t, err)
		} else {
			assert.Nil(t, err)
		}
		assert.Equal(t, test.outputBuf, test.inputBuf)
		assert.Equal(t, test.outputN, n)
	}
}

func TestParseHave(t *testing.T) {
	tests := map[string]struct {
		input  *Message
		output int
		fails  bool
	}{
		"parse valid message": {
			input:  &Message{ID: MsgHave, Payload: []byte{0x00, 0x00, 0x00, 0x04}},
			output: 4,
			fails:  false,
		},
		"wrong message type": {
			input:  &Message{ID: MsgPiece, Payload: []byte{0x00, 0x00, 0x00, 0x04}},
			output: 0,
			fails:  true,
		},
		"payload too short": {
			input:  &Message{ID: MsgHave, Payload: []byte{0x00, 0x00, 0x04}},
			output: 0,
			fails:  true,
		},
		"payload too long": {
			input:  &Message{ID: MsgHave, Payload: []byte{0x00, 0x00, 0x00, 0x00, 0x04}},
			output: 0,
			fails:  true,
		},
	}

	for _, test := range tests {
		index, err := ParseHave(test.input)
		if test.fails {
			assert.NotNil(t, err)
		} else {
			assert.Nil(t, err)
		}
		assert.Equal(t, test.output, index)
	}
}

func TestHasPiece(t *testing.T) {
	bf := Bitfield{0b01010100, 0b01010100}
	output := []bool{false, true, false, true, false, true, false, false, false, true, false, true, false, true, false, false}
	for i := 0; i < len(output); i++ {
		assert.Equal(t, output[i], bf.HasPiece(i))
	}
}

func TestSetPiece(t *testing.T) {
	tests := []struct {
		input  Bitfield
		index  int
		output Bitfield
	}{
		{
			input:  Bitfield{0b01010100, 0b01010100},
			index:  4,
			output: Bitfield{0b01011100, 0b01010100},
		},
		{
			input:  Bitfield{0b01010100, 0b01010100},
			index:  9,
			output: Bitfield{0b01010100, 0b01010100},
		},
		{
			input:  Bitfield{0b01010100, 0b01010100},
			index:  15,
			output: Bitfield{0b01010100, 0b01010101},
		},
	}

	for _, test := range tests {
		bf := test.input
		bf.SetPiece(test.index)
		assert.Equal(t, test.output, bf)
	}
}
