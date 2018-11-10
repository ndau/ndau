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
var hashToHeightSearchKeyPrefix = "h:"

// Client is a search Client that implements IncrementalIndexer.
type Client struct {
	*metasearch.Client

	blockHash string
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

	search.blockHash = ""
	search.blockHeight = 0
	search.maxHeight = 0

	return search, nil
}

func formatHashToHeightSearchKey(hash string) string {
	return hashToHeightSearchKeyPrefix + strings.ToLower(hash)
}

func (search *Client) onIndexingComplete() (updateCount int, insertCount int, err error) {
	searchKey := formatHashToHeightSearchKey(search.blockHash)
	searchValue := fmt.Sprintf("%d", search.blockHeight)

	updateCount, insertCount, err = search.indexKeyValue(searchKey, searchValue)
	if err != nil {
		return updateCount, insertCount, err
	}

	// No need to keep this data around any longer.
	search.blockHash = ""
	search.blockHeight = 0

	// Save this off so the next initial scan will only go this far.
	search.Client.SetHeight(search.maxHeight)

	return updateCount, insertCount, nil
}

// Index a single key-value pair at the given height.
func (search *Client) indexKeyValue(
	searchKey string,
	searchValue string,
) (updateCount int, insertCount int, err error) {
	existingValue, err := search.Client.Get(searchKey)
	if err != nil {
		return 0, 0, err
	}

	err = search.Client.Set(searchKey, searchValue)
	if err != nil {
		return 0, 0, err
	}

	if existingValue == "" {
		updateCount = 0
		insertCount = 1
	} else {
		updateCount = 1
		insertCount = 0
	}

	return updateCount, insertCount, nil
}
