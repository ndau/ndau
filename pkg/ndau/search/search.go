package search

// The public API for searching the index.

import (
	"strconv"
)

// SearchBlockHash returns the height of the given block hash.
// Returns 0 and no error if the given block hash was not found in the index.
func (search *Client) SearchBlockHash(blockHash string) (uint64, error) {
	searchKey := formatBlockHashToHeightSearchKey(blockHash)

	searchValue, err := search.Client.Get(searchKey)
	if err != nil {
		return 0, err
	}

	if searchValue == "" {
		// Empty search results.  No error.
		return 0, nil
	}

	return strconv.ParseUint(searchValue, 10, 64)
}

// SearchTxHash returns the height of block containing the given tx hash.
// It also returns the transaction offset within the block.
// Returns 0, 0 and no error if the given tx hash was not found in the index.
func (search *Client) SearchTxHash(txHash string) (uint64, int, error) {
	searchKey := formatTxHashToHeightSearchKey(txHash)

	searchValue, err := search.Client.Get(searchKey)
	if err != nil {
		return 0, 0, err
	}

	if searchValue == "" {
		// Empty search results.  No error.
		return 0, 0, nil
	}

	valueData := TxValueData{}
	err = valueData.Unmarshal(searchValue)
	if err != nil {
		return 0, 0, err
	}

	return valueData.BlockHeight, valueData.TxOffset, nil
}
