// Package magnetlink provides support for parsing [magnet links]
//
// [magnet links]: https://en.wikipedia.org/wiki/Magnet_URI_scheme
package magnetlink

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/pasztorpisti/qs"
	"github.com/sauromates/leech/internal/bthash"
)

// MagnetLink represents parsed info from an actual magnet link.
// @todo convert to a torrent
type MagnetLink struct {
	InfoHash   bthash.Hash `qs:"xt,req"`
	Name       string      `qs:"dn,req"`
	Length     int         `qs:"xl,opt"`
	TrackerURL string      `qs:"tr,opt"`
}

// Unmarshal parses magnet link into a struct using [qs.UnmarshalValues] with
// some additional logic (i.e. retrieving info hash from complex string).
func Unmarshal(link string, target any) error {
	query, ok := strings.CutPrefix(link, "magnet:?")
	if !ok {
		return fmt.Errorf("[ERROR] %s is not a valid magnet link", link)
	}

	params, err := url.ParseQuery(query)
	if err != nil {
		return err
	}

	ih := params.Get("xt")
	if btih, ok := strings.CutPrefix(ih, "urn:btih:"); ok {
		params.Set("xt", btih)
	}

	if err := qs.UnmarshalValues(target, params); err != nil {
		return err
	}

	return nil
}
