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
const sysvarKeyToValueSearchKeyPrefix = "s:"
const txHashToHeightSearchKeyPrefix = "t:"
const unionPrefix = "u:"
const txTypeToHeightSearchKeyPrefix = "y:"

// This is used to be able to give transactions a float64 score in a sorted set where the integer
// part of the score is the block height, and the fractional part contains the tx offset within
// the block.  Typically there are zero or one transactions in a block.  If we ever had anything
// close to this many transactions in a block, it will be a good problem to have.
// For example, the 3rd transaction (tx offset = 2) in block 10 would have a score of 10.002.
// Float determinism is not a concern here since we just want each transaction to have a unique
// and well-defined order on the blockchain when compared to other transactions.
const maxTxsPerBlock = 1000

// AppIndexable is an app which can help index its transactions.
//
// It's really only a thing in order to avoid circular imports; it will always
// in actuality be an ndau.App
type AppIndexable interface {
	GetAccountAddresses(tx metatx.Transactable) ([]string, error)
	GetState() metastate.State
	CalculateTxFeeNapu(tx metatx.Transactable) (uint64, error)
	CalculateTxSIBNapu(tx metatx.Transactable) (uint64, error)
}

// SysvarIndexable is a Transactable that has sysar data that we want to index.
type SysvarIndexable interface {
	metatx.Transactable

	// We use separate methods (instead of a struct to house the data) to avoid extra memory use.
	GetName() string
	GetValue() []byte
}

// Client is a search Client that implements IncrementalIndexer.
type Client struct {
	*metasearch.Client

	// Used when collecting sysvar keys to index.  In the case of initial indexing,
	// this combines keys and values over possibly multiple blocks.
	sysvarKeyToValueData map[string]*ValueData

	// Used for getting account data to index.
	app AppIndexable

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
func NewClient(address string, version int, app AppIndexable) (search *Client, err error) {
	search = &Client{}
	search.Client, err = metasearch.NewClient(address, version)
	if err != nil {
		return nil, err
	}

	search.sysvarKeyToValueData = nil
	search.app = app
	search.txs = nil
	search.blockTime = time.Time{}
	search.blockHash = ""
	search.blockHeight = 0
	search.nextHeight = 0

	return search, nil
}

// Helper function for generating unique search sysvar keys within the redis database.
func formatSysvarKeyToValueSearchKey(key string) string {
	return sysvarKeyToValueSearchKeyPrefix + key
}

func formatBlockHashToHeightSearchKey(hash string) string {
	return blockHashToHeightSearchKeyPrefix + strings.ToLower(hash)
}

func formatTxHashToHeightSearchKey(hash string) string {
	return txHashToHeightSearchKeyPrefix + hash
}

func formatUniqueUnionSearchKey() string {
	// Use time now to effectively guarantee uniqueness for every caller.
	return fmt.Sprintf("%s%d", unionPrefix, time.Now().UnixNano())
}

func formatTxTypeToHeightSearchKey(typeName string) string {
	return txTypeToHeightSearchKeyPrefix + typeName
}

func formatAccountAddressToHeightSearchKey(addr string) string {
	return accountAddressToHeightSearchKeyPrefix + addr
}

// Index all the key-value pairs in the search's sysvarKeyToValueData mapping, then clear the map.
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
	for searchKey, data := range search.sysvarKeyToValueData {
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

					height := valueData.Height
					valueBase64 := valueData.ValueBase64

					if !hasValue || dupeHeight < height && height <= data.Height {
						dupeValueBase64 = valueBase64
						dupeHeight = height
						hasValue = true
						if dupeHeight == data.Height {
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

			if hasValue && dupeValueBase64 == data.ValueBase64 {
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
		search.Client.IndexDateToHeight(search.blockTime, search.nextHeight-1)
	updateCount += updCount
	insertCount += insCount
	if err != nil {
		return updateCount, insertCount, err
	}

	// No need to keep this data around any longer.
	search.sysvarKeyToValueData = nil
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

// Index a single key-value pair into a sorted set.
func (search *Client) indexTxType(txType, txHash string, blockHeight uint64, txOffset int) (
	updateCount int, insertCount int, err error,
) {
	if txOffset < 0 || txOffset >= maxTxsPerBlock {
		// If this happens, we either have to increase maxTxsPerBlock or compute height-txoffset
		// scores in a different way that doesn't have this limitation.  Either way, a full re-
		// index will be necessary once solved.
		return 0, 0, fmt.Errorf("Tx offset out of range: %d >= %d", txOffset, maxTxsPerBlock)
	}

	searchKey := formatTxTypeToHeightSearchKey(txType)
	score := float64(blockHeight) + float64(txOffset)/float64(maxTxsPerBlock)
	count, err := search.Client.ZAdd(searchKey, score, txHash)
	if err != nil {
		return 0, 0, err
	}

	if count == 0 {
		updateCount = 1
		insertCount = 0
	} else {
		updateCount = 0
		insertCount = int(count) // count == 1 for a single ZADD.
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

		searchKey := formatSysvarKeyToValueSearchKey(key)

		// Detect the first time we've encountered this key.
		data, hasValue := search.sysvarKeyToValueData[searchKey]
		if !hasValue {
			search.sysvarKeyToValueData[searchKey] = &ValueData{
				Height:      search.blockHeight,
				ValueBase64: valueBase64,
			}
			continue
		}

		// Skip indexing adjacent blocks having the same value for the given key.
		// This assumes we're iterating blocks in order from the head to genesis.
		if data.ValueBase64 == valueBase64 {
			// Save off the current height of the iteration.  We do this when we're
			// not indexing it so we eventually index with the lowest block height
			// seen for a given search key.
			data.Height = search.blockHeight
			continue
		}

		// This is only a sanity check.  Noms doesn't preserve any but the last-set value
		// for a given key and height.  So we'll never encounter this case.
		if data.Height == search.blockHeight {
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
		data.Height = search.blockHeight
		data.ValueBase64 = valueBase64
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
	valueData := TxValueData{BlockHeight: search.blockHeight}
	acctValueData := AccountTxValueData{search.blockHeight, 0, 0}

	for txOffset, tx := range search.txs {
		// Index transaction hash.
		txHash := metatx.Hash(tx)
		searchKey := formatTxHashToHeightSearchKey(txHash)
		valueData.TxOffset = txOffset
		valueData.Fee, err = search.app.CalculateTxFeeNapu(tx)
		if err != nil {
			return updateCount, insertCount, err
		}
		valueData.SIB, err = search.app.CalculateTxSIBNapu(tx)
		if err != nil {
			return updateCount, insertCount, err
		}
		acctValueData.TxOffset = txOffset
		searchValue := valueData.Marshal()

		updCount, insCount, err := search.indexKeyValue(searchKey, searchValue)
		updateCount += updCount
		insertCount += insCount
		if err != nil {
			return updateCount, insertCount, err
		}

		// Update the appropriate tx type index.
		txType := metatx.NameOf(tx)
		updCount, insCount, err = search.indexTxType(txType, txHash, search.blockHeight, txOffset)
		updateCount += updCount
		insertCount += insCount
		if err != nil {
			return updateCount, insertCount, err
		}

		// Index account addresses associated with the transaction.
		addresses, err := search.app.GetAccountAddresses(tx)
		if err != nil {
			err = fmt.Errorf(
				"tx with hash %s attempted to get account addresses: %s",
				metatx.Hash(tx),
				err.Error(),
			)
			return updateCount, insertCount, err
		}

		for _, addr := range addresses {
			acct, hasAccount := search.app.GetState().(*backing.State).Accounts[addr]
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

	return updateCount, insertCount, nil
}
