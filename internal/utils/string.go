package utils

import (
	"encoding/hex"

	"github.com/pasztorpisti/qs"
)

// An alias for `utils.BTString` type used all over as default
// length of BitTorrent information strings like `info_hash`
// or `peer_id`
type BTString [20]byte

// MarshalQS makes [BTString] implement [qs.MarshalQS] interface.
func (bts BTString) MarshalQS(opts *qs.MarshalOptions) ([]string, error) {
	return []string{hex.EncodeToString(bts[:])}, nil
}

// UnmarshalQS makes [BTString] implement [qs.UnmarshalQS] interface.
func (bts *BTString) UnmarshalQS(a []string, opts *qs.UnmarshalOptions) error {
	s, err := opts.SliceToString(a)
	if err != nil {
		return err
	}

	*bts = DecodeHEX(s)

	return nil
}

// DecodeHEX attempts to convert hexadecimal string to [20]byte slice.
// In case of error empty slice is returned.
func DecodeHEX(s string) BTString {
	hashBytes, err := hex.DecodeString(s)
	if err != nil {
		return [20]byte{}
	}

	if len(hashBytes) != 20 {
		return [20]byte{}
	}

	var out BTString
	copy(out[:], hashBytes)

	return out
}
