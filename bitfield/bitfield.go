package bitfield

// A Bitfield represents the pieces that a peer has.
// It's a data structure that peers use to efficiently encode which pieces
// they are able to send us.
type Bitfield []byte

// HasPiece tells if a Bitfield has a particular index
func (b Bitfield) HasPiece(index int) bool {
	// Bitfield looks like a byte array (grid).
	// You can think of it like a coffee shop loyalty card. We start with
	// a blank card of all 0, and flip bits to 1 to mark
	// their positions as "stamped".
	byteIndex := index / 8 // which row in grid
	offset := index % 8    // which column in grid

	return b[byteIndex]>>(7-offset)&1 != 0 // bitwise manipulation
}

// SetPiece sets a bit in the Bitfield
func (b Bitfield) SetPiece(index int) {
	byteIndex := index / 8
	offset := index % 8
	b[byteIndex] |= 1 << (7 - offset)
}
