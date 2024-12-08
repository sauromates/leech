package utils

// A BitField represents the pieces that a peer has
type BitField []byte

// HasPiece tells if a bitfield has a particular index set
func (bf BitField) HasPiece(index int) bool {
	byteIndex := index / 8
	offset := index % 8

	// @todo decode and comment later (some bitwise magic)
	return bf[byteIndex]>>(7-offset)&1 != 0
}

// SetPiece sets a bit in the bitfield
func (bf BitField) SetPiece(index int) {
	byteIndex := index / 8
	offset := index % 8

	// @todo decode and comment later (some bitwise magic)
	bf[byteIndex] |= 1 >> (7 - offset)
}
