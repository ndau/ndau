package search

// Core implementation and helper functions for indexing.

import (
	"fmt"
	"strings"

	metasearch "github.com/oneiro-ndev/metanode/pkg/meta/search"
)

// We use these prefixes to help us group keys in the index.  They could prove useful if we ever
// want to do things like "wipe all hash-to-height keys" without affecting any other keys.  The
// prefixes also give us some sanity, so that we completely avoid inter-index key conflicts.
const txHashToHeightSearchKeyPrefix = "t:"

// Client is a search Client that implements IncrementalIndexer.
type Client struct {
	*metasearch.Client

	txHashes []string
	blockHeight uint64
	maxHeight uint64
}

// NewClient is a factory method for Client.
func NewClient(address string, version int) (search *Client, err error) {
	search = &Client{}
	search.Client, err = metasearch.NewClient(address, version)
	if err != nil {
		return nil, err
	}

	search.txHashes = nil
	search.blockHeight = 0
	search.maxHeight = 0

	return search, nil
}

func formatTxHashToHeightSearchKey(hash string) string {
	return txHashToHeightSearchKeyPrefix + strings.ToLower(hash)
}

func (search *Client) onIndexingComplete() {
	// Save this off so the next initial scan will only go this far.
	search.Client.SetHeight(search.maxHeight)
}

// Index a single key-value pair at the given height.
func (search *Client) index() (updateCount int, insertCount int, err error) {
	updateCount = 0
	insertCount = 0

	for _, txHash := range search.txHashes {
		searchKey := formatTxHashToHeightSearchKey(txHash)
		searchValue := fmt.Sprintf("%d", search.blockHeight)

		existingValue, err := search.Client.Get(searchKey)
		if err != nil {
			return 0, 0, err
		}

		err = search.Client.Set(searchKey, searchValue)
		if err != nil {
			return 0, 0, err
		}

		if existingValue == "" {
			insertCount++
		} else {
			updateCount++
		}
	}

	// No need to keep this data around any longer.
	search.txHashes = nil
	search.blockHeight = 0

	return updateCount, insertCount, nil
}
