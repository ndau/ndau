package search

// The public API for searching the index.

import (
	"strconv"
)

// SearchBlockHeightByBlockHash returns the height of the given block hash.
// Returns 0 and no error if the given block hash was not found in the index.
func (search *Client) SearchBlockHeightByBlockHash(hash string) (uint64, error) {
	searchKey := formatHashToHeightSearchKey(hash)

	value, err := search.Client.Get(searchKey)
	if err != nil {
		return 0, err
	}

	if value == "" {
		return 0, nil
	}

	return strconv.ParseUint(value, 10, 64)
}
