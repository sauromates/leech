// Package bthash provides a utility type [Hash] for working with BitTorrent
// identifiers in a standardized manner.
package bthash

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"io"
	"strings"

	"github.com/pasztorpisti/qs"
)

// ErrEmptyHash is returned when either string can't be decoded into [Hash]
// or decoded string is not of length 20 (which is required for [20]byte) hash.
var ErrEmptyHash error = errors.New("[ERROR] Decoding error, hash is empty")

// randomizer is used for creating random hashes
var randomizer io.Reader = rand.Reader

// Hash is an alias for [20]byte array used for BitTorrent-related identifiers
// such as info hashes, peer IDs, node IDs, etc.
//
// Hash also implements [qs.MarshalQS] and [qs.UnmarshalQS] interfaces,
// enabling its use for parsing and serializing magnet links.
type Hash [20]byte

// NewFromString decodes hexadecimal string value into a [20]byte array.
// Any errors would lead to empty [Hash] returned.
func NewFromString(s string) Hash {
	b, err := hex.DecodeString(s)
	if err != nil || len(b) != 20 {
		return Hash{}
	}

	var bth Hash
	copy(bth[:], b)

	return bth
}

// NewRandom creates [20]byte hash via randomizer which is set at package level.
//
// Return value is an empty hash if the randomizer fails with error.
func NewRandom() Hash {
	var hash Hash
	if _, err := randomizer.Read(hash[:]); err != nil {
		return Hash{}
	}

	return hash
}

// UnmarshalQS creates new [Hash] from query string params by implementing
// [qs.UnmarshalQS] interface.
func (bth *Hash) UnmarshalQS(a []string, opts *qs.UnmarshalOptions) error {
	s, err := opts.SliceToString(a)
	if err != nil {
		return err
	}

	*bth = NewFromString(s)

	if *bth == (Hash{}) {
		return ErrEmptyHash
	}

	return nil
}

// MarshalQS returns a slice of strings with single hexadecimal value of
// [Hash] thus implementing [qs.MarshalQS] interface.
func (bth Hash) MarshalQS(opts *qs.MarshalOptions) ([]string, error) {
	if bth == (Hash{}) {
		return []string{}, ErrEmptyHash
	}

	return []string{bth.String()}, nil
}

// String returns hexadecimal encoded BitTorrent hash. Empty hash produces
// empty string instead of 20 zeroes.
func (bth Hash) String() string {
	if bth == (Hash{}) {
		return ""
	}

	return strings.ToUpper(hex.EncodeToString(bth[:]))
}
