package search

// Core implementation and helper functions for indexing.

import (
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	metasearch "github.com/oneiro-ndev/metanode/pkg/meta/search"
	metastate "github.com/oneiro-ndev/metanode/pkg/meta/state"
	metatx "github.com/oneiro-ndev/metanode/pkg/meta/transaction"
	"github.com/oneiro-ndev/ndau/pkg/ndau/backing"
)

// We use these prefixes to help us group keys in the index.  They could prove useful if we ever
// want to do things like "wipe all hash-to-height keys" without affecting any other keys.  The
// prefixes also give us some sanity, so that we completely avoid inter-index key conflicts.
// NOTE: These must not conflict with dateRangeToHeightSearchKeyPrefix defined in metanode.
const accountAddressToHeightSearchKeyPrefix = "a:"
const blockHashToHeightSearchKeyPrefix = "b:"
const keyToValueSearchKeyPrefix = "k:"
const txHashToHeightSearchKeyPrefix = "t:"

// KeyValueIndexable is a Transactable that has key-value data that we want to index.
type KeyValueIndexable interface {
	metatx.Transactable

	// We use separate methods (instead of a struct to house the data) to avoid extra memory use.
	GetKey() string
	GetValue() []byte
}

// AddressIndexable is a Transactable that has addresses associated with it that we want to index.
type AddressIndexable interface {
	metatx.Transactable

	// We use separate methods (instead of a struct to house the data) to avoid extra memory use.
	GetAccountAddresses() []string
}

// Client is a search Client that implements IncrementalIndexer.
type Client struct {
	*metasearch.Client

	// Used when collecting keys to index.  In the case of initial indexing,
	// this combines keys and values over possibly multiple blocks.
	keyToValueData map[string]*ValueData

	// Used for getting account data to index.
	state metastate.State

	// Used for indexing transaction hashes.
	txs []metatx.Transactable

	// Used for indexing the block hash at the current height.
	blockHash string

	// These pertain to the current block we're indexing.
	blockTime   time.Time
	blockHeight uint64

	// The next height we will index after the current incremental/initial indexing completes.
	nextHeight uint64
}

// NewClient is a factory method for Client.
func NewClient(address string, version int) (search *Client, err error) {
	search = &Client{}
	search.Client, err = metasearch.NewClient(address, version)
	if err != nil {
		return nil, err
	}

	search.keyToValueData = nil
	search.state = nil
	search.txs = nil
	search.blockTime = time.Time{}
	search.blockHash = ""
	search.blockHeight = 0
	search.nextHeight = 0

	return search, nil
}

// Helper function for generating unique search keys within the redis database.
func formatKeyToValueSearchKey(key string) string {
	return keyToValueSearchKeyPrefix + key
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

// Index a all the key-value pairs in the search's keyToValueData mapping, then clear the map.
// checkForDupes is used for merging any duplicate keys we find in the mapping.
func (search *Client) onIndexingComplete(
	checkForDupes bool,
) (updateCount int, insertCount int, err error) {
	updateCount = 0
	insertCount = 0

	// We'll reuse this for unmarshaling data into it.
	valueData := &ValueData{}

	// When we initially index, we only indexed when we noticed a change in a given key's value.
	// After we've completed the blockchain crawl, whatever's left is the first occurrence of a
	// given key at its latest value.  So we index them at that point here.
	// In the case of incremental indexing, we fill the mapping with the new/changed values and
	// index them all here when the block is committed.
	for searchKey, data := range search.keyToValueData {
		skip := false

		if checkForDupes {
			// Find the potential dupe value for this key in the index.
			hasValue := false
			dupeHeight := uint64(0)
			dupeValueBase64 := ""

			err = search.Client.SScan(searchKey,
				func(searchValue string) error {
					err := valueData.Unmarshal(searchValue)
					if err != nil {
						return err
					}

					height := valueData.height
					valueBase64 := valueData.valueBase64

					if !hasValue || dupeHeight < height && height <= data.height {
						dupeValueBase64 = valueBase64
						dupeHeight = height
						hasValue = true
						if dupeHeight == data.height {
							// Found potential dupe at the right height.
							// No need to iterate further.
							return metastate.StopIteration()
						}
					}
					return nil
				})
			if err != nil && !metastate.IsStopIteration(err) {
				return updateCount, insertCount, err
			}

			if hasValue && dupeValueBase64 == data.valueBase64 {
				skip = true
			}
		}

		if !skip {
			updCount, insCount, err := search.indexKeyValueWithHistory(searchKey, data.Marshal())
			updateCount += updCount
			insertCount += insCount
			if err != nil {
				return updateCount, insertCount, err
			}
		}
	}

	// Index date to height as needed.
	updCount, insCount, err :=
		search.Client.IndexDateToHeight(search.blockTime, search.nextHeight - 1)
	updateCount += updCount
	insertCount += insCount
	if err != nil {
		return updateCount, insertCount, err
	}

	// No need to keep this data around any longer.
	search.keyToValueData = nil
	search.state = nil
	search.txs = nil
	search.blockTime = time.Time{}
	search.blockHash = ""
	search.blockHeight = 0

	// Save this off so the next initial scan will only go this far.
	search.Client.SetNextHeight(search.nextHeight)

	return updateCount, insertCount, nil
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

// Index all the sysvar key-value pairs in the given state at the current search.blockHeight.
func (search *Client) indexState(
	st *backing.State,
) (updateCount int, insertCount int, err error) {
	updateCount = 0
	insertCount = 0

	for key, value := range st.Sysvars {
		valueBase64 := base64.StdEncoding.EncodeToString(value)

		searchKey := formatKeyToValueSearchKey(key)

		// Detect the first time we've encountered this key.
		data, hasValue := search.keyToValueData[searchKey]
		if !hasValue {
			search.keyToValueData[searchKey] = &ValueData{
				height:      search.blockHeight,
				valueBase64: valueBase64,
			}
			continue
		}

		// Skip indexing adjacent blocks having the same value for the given key.
		// This assumes we're iterating blocks in order from the head to genesis.
		if data.valueBase64 == valueBase64 {
			// Save off the current height of the iteration.  We do this when we're
			// not indexing it so we eventually index with the lowest block height
			// seen for a given search key.
			data.height = search.blockHeight
			continue
		}

		// This is only a sanity check.  Noms doesn't preserve any but the last-set value
		// for a given key and height.  So we'll never encounter this case.
		if data.height == search.blockHeight {
			continue
		}
		
		// Index the old value and height since we just found the block where the value
		// changed.  The caller will index the value when it was originally set.
		updCount, insCount, err := search.indexKeyValueWithHistory(searchKey, data.Marshal())
		updateCount += updCount
		insertCount += insCount
		if err != nil {
			return updateCount, insertCount, err
		}

		// Save off the current value of the iteration.  We'll eventually index it at the
		// lowest height we see for it.
		data.height = search.blockHeight
		data.valueBase64 = valueBase64
	}

	return updateCount, insertCount, nil
}

// Index everything we have in the Client at the current search.blockHeight.
func (search *Client) index() (updateCount int, insertCount int, err error) {
	updateCount = 0
	insertCount = 0

	// If we have no block hash, we have nothing to index.
	if search.blockHash == "" {
		return updateCount, insertCount, nil
	}

	heightValue := fmt.Sprintf("%d", search.blockHeight)

	blockHashKey := formatBlockHashToHeightSearchKey(search.blockHash)
	updCount, insCount, err := search.indexKeyValue(blockHashKey, heightValue)
	updateCount += updCount
	insertCount += insCount
	if err != nil {
		return updateCount, insertCount, err
	}

	// We'll reuse these for marshaling data into it.
	valueData := TxValueData{search.blockHeight, 0}
	acctValueData := AccountTxValueData{search.blockHeight, 0, 0}

	for txOffset, tx := range search.txs {
		// Index transaction hash.
		txHash := metatx.Hash(tx)
		searchKey := formatTxHashToHeightSearchKey(txHash)
		valueData.TxOffset = txOffset
		acctValueData.TxOffset = txOffset
		searchValue := valueData.Marshal()

		updCount, insCount, err := search.indexKeyValue(searchKey, searchValue)
		updateCount += updCount
		insertCount += insCount
		if err != nil {
			return updateCount, insertCount, err
		}

		// Index account addresses associated with the transaction.
		indexable, isIndexable := tx.(AddressIndexable)
		if isIndexable {
			addresses := indexable.GetAccountAddresses()

			for _, addr := range addresses {
				acct, hasAccount := search.state.(*backing.State).Accounts[addr]
				if hasAccount {
					searchKey := formatAccountAddressToHeightSearchKey(addr)
					acctValueData.Balance = acct.Balance
					searchValue := acctValueData.Marshal()

					updCount, insCount, err :=
						search.indexKeyValueWithHistory(searchKey, searchValue)
					updateCount += updCount
					insertCount += insCount
					if err != nil {
						return updateCount, insertCount, err
					}
				}
			}
		}
	}

	return updateCount, insertCount, nil
}
