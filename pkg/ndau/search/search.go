package search

// The public API for searching the index.

import (
	"strconv"
)

// SearchBlockHeightByTxHash returns the height of block containing the given tx hash.
// Returns 0 and no error if the given tx hash was not found in the index.
func (search *Client) SearchBlockHeightByTxHash(hash string) (uint64, error) {
	searchKey := formatTxHashToHeightSearchKey(hash)

	value, err := search.Client.Get(searchKey)
	if err != nil {
		return 0, err
	}

	if value == "" {
		return 0, nil
	}

	return strconv.ParseUint(value, 10, 64)
}
