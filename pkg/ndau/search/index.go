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
const blockHashToHeightSearchKeyPrefix = "b:"
const txHashToHeightSearchKeyPrefix = "t:"

// Client is a search Client that implements IncrementalIndexer.
type Client struct {
	*metasearch.Client

	txHashes []string
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

	search.txHashes = nil
	search.blockHash = ""
	search.blockHeight = 0
	search.maxHeight = 0

	return search, nil
}

func formatBlockHashToHeightSearchKey(hash string) string {
	return blockHashToHeightSearchKeyPrefix + strings.ToLower(hash)
}

func formatTxHashToHeightSearchKey(hash string) string {
	return txHashToHeightSearchKeyPrefix + strings.ToLower(hash)
}

func (search *Client) onIndexingComplete() {
	// Save this off so the next initial scan will only go this far.
	search.Client.SetHeight(search.maxHeight)
}

// Index a single key-value pair.
func (search *Client) indexKeyValue(searchKey, searchValue string) (
	updateCount int, insertCount int, err error,
) {
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

// Index everything we have in the Client.
func (search *Client) index() (updateCount int, insertCount int, err error) {
	updateCount = 0
	insertCount = 0

	heightValue := fmt.Sprintf("%d", search.blockHeight)

	blockHashKey := formatBlockHashToHeightSearchKey(search.blockHash)
	updCount, insCount, err := search.indexKeyValue(blockHashKey, heightValue)
	updateCount += updCount
	insertCount += insCount
	if err != nil {
		return updateCount, insertCount, err
	}

	// We'll reuse this for marshaling data into it.
	valueData := TxValueData{search.blockHeight, 0}

	for txOffset, txHash := range search.txHashes {
		txHashKey := formatTxHashToHeightSearchKey(txHash)
		valueData.TxOffset = uint64(txOffset)
		searchValue := valueData.Marshal()

		updCount, insCount, err := search.indexKeyValue(txHashKey, searchValue)
		updateCount += updCount
		insertCount += insCount
		if err != nil {
			return updateCount, insertCount, err
		}
	}

	// No need to keep this data around any longer.
	search.txHashes = nil
	search.blockHash = ""
	search.blockHeight = 0

	return updateCount, insertCount, nil
}
