package search

// Core implementation and helper functions for indexing.

import (
	"fmt"
	"strings"

	metasearch "github.com/oneiro-ndev/metanode/pkg/meta/search"
	metatx "github.com/oneiro-ndev/metanode/pkg/meta/transaction"
)

// We use these prefixes to help us group keys in the index.  They could prove useful if we ever
// want to do things like "wipe all hash-to-height keys" without affecting any other keys.  The
// prefixes also give us some sanity, so that we completely avoid inter-index key conflicts.
const blockHashToHeightSearchKeyPrefix = "b:"
const txHashToHeightSearchKeyPrefix = "t:"
const accountAddressToHeightSearchKeyPrefix = "a:"

// Indexable is an indexable Transactable.
type Indexable interface {
	metatx.Transactable

	// We use separate methods (instead of a struct to house the data) to avoid extra memory use.
	GetAccountAddresses() []string
}

// Client is a search Client that implements IncrementalIndexer.
type Client struct {
	*metasearch.Client

	txs []metatx.Transactable
	blockHash string
	blockHeight uint64
	nextHeight uint64
}

// NewClient is a factory method for Client.
func NewClient(address string, version int) (search *Client, err error) {
	search = &Client{}
	search.Client, err = metasearch.NewClient(address, version)
	if err != nil {
		return nil, err
	}

	search.txs = nil
	search.blockHash = ""
	search.blockHeight = 0
	search.nextHeight = 0

	return search, nil
}

func formatBlockHashToHeightSearchKey(hash string) string {
	return blockHashToHeightSearchKeyPrefix + strings.ToLower(hash)
}

func formatTxHashToHeightSearchKey(hash string) string {
	return txHashToHeightSearchKeyPrefix + hash
}

func formatAccountAddressToHeightSearchKey(addr string) string {
	return accountAddressToHeightSearchKeyPrefix + addr
}

func (search *Client) onIndexingComplete() {
	// No need to keep this data around any longer.
	search.txs = nil
	search.blockHash = ""
	search.blockHeight = 0

	// Save this off so the next initial scan will only go this far.
	search.Client.SetNextHeight(search.nextHeight)
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

// Index a single key-value pair with history.
func (search *Client) indexKeyValueWithHistory(searchKey, searchValue string) (
	updateCount int, insertCount int, err error,
) {
	count, err := search.Client.SAdd(searchKey, searchValue)
	if err != nil {
		return 0, 0, err
	}

	if count == 0 {
		updateCount = 1
		insertCount = 0
	} else {
		updateCount = 0
		insertCount = int(count) // count == 1 for a single SADD.
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

	for txOffset, tx := range search.txs {
		// Index transaction hash.
		txHash := metatx.Hash(tx)
		searchKey := formatTxHashToHeightSearchKey(txHash)
		valueData.TxOffset = txOffset
		searchValue := valueData.Marshal()

		updCount, insCount, err := search.indexKeyValue(searchKey, searchValue)
		updateCount += updCount
		insertCount += insCount
		if err != nil {
			return updateCount, insertCount, err
		}

		// Index account addresses associated with the transaction.
		indexable, isIndexable := tx.(Indexable)
		if isIndexable {
			addresses := indexable.GetAccountAddresses()

			for _, addr := range addresses {
				searchKey := formatAccountAddressToHeightSearchKey(addr)

				updCount, insCount, err := search.indexKeyValueWithHistory(searchKey, searchValue)
				updateCount += updCount
				insertCount += insCount
				if err != nil {
					return updateCount, insertCount, err
				}
			}
		}
	}

	return updateCount, insertCount, nil
}
