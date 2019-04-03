package search

// The public API for searching the index.

import (
	"encoding/base64"
	"sort"
	"strconv"

	srch "github.com/oneiro-ndev/metanode/pkg/meta/search"
	"github.com/oneiro-ndev/ndau/pkg/query"
)

// SearchSysvarHistory returns value history for the given sysvar using an index under the hood.
// The response is sorted by ascending block height, each entry is where the key's value changed.
func (search *Client) SearchSysvarHistory(
	sysvar string, pageIndex int, pageSize int,
) (khr *query.SysvarHistoryResponse, err error) {
	khr = new(query.SysvarHistoryResponse)

	searchKey := formatSysvarKeyToValueSearchKey(sysvar)

	// We'll reuse this for unmarshaling data into it.
	valueData := &ValueData{}

	err = search.Client.SScan(searchKey, func(searchValue string) error {
		err := valueData.Unmarshal(searchValue)
		if err != nil {
			return err
		}

		height := valueData.height
		valueBase64 := valueData.valueBase64

		value, err := base64.StdEncoding.DecodeString(valueBase64)
		if err != nil {
			return err
		}

		khr.History = append(khr.History, query.SysvarHistoricalValue{
			Height: height,
			Value:  value,
		})

		return nil
	})

	// Sort by ascending height.  Even if we had used ZAdd() with ZScan(), we'd still have to sort
	// because of the way it returns pages of results under the hood that we'd have to merge.
	sort.Slice(khr.History, func(i, j int) bool {
		return khr.History[i].Height < khr.History[j].Height
	})

	// Reduce the full results list down to the requested page.  There is some wasted effort with
	// this approach, but we support the worst case, which is to return all results.  In practice,
	// getting the full list from the underlying index is fast, with tolerable sorting speed.
	offsetStart, offsetEnd := srch.GetPageOffsets(pageIndex, pageSize, len(khr.History))
	khr.History = khr.History[offsetStart:offsetEnd]

	return khr, err
}

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

// SearchAccountHistory returns an array of block height and txoffset pairs associated with the
// given account address.
func (search *Client) SearchAccountHistory(
	addr string, pageIndex int, pageSize int,
) (ahr *AccountHistoryResponse, err error) {
	ahr = new(AccountHistoryResponse)

	searchKey := formatAccountAddressToHeightSearchKey(addr)

	err = search.Client.SScan(searchKey, func(searchValue string) error {
		valueData := AccountTxValueData{}
		err := valueData.Unmarshal(searchValue)
		if err != nil {
			return err
		}

		ahr.Txs = append(ahr.Txs, valueData)

		return nil
	})

	// Sort by ascending height, with ascending tx offset as the secondary sort order.
	// Even if we had used ZAdd() with ZScan(), we'd still have to sort because of the way it
	// returns pages of results under the hood that we'd have to merge.
	sort.Slice(ahr.Txs, func(i, j int) bool {
		txi := &ahr.Txs[i]
		txj := &ahr.Txs[j]
		return txi.BlockHeight < txj.BlockHeight ||
			txi.BlockHeight == txj.BlockHeight && txi.TxOffset < txj.TxOffset
	})

	// Reduce the full results list down to the requested page.  There is some wasted effort with
	// this approach, but we support the worst case, which is to return all results.  In practice,
	// getting the full list from the underlying index is fast, with tolerable sorting speed.
	offsetStart, offsetEnd := srch.GetPageOffsets(pageIndex, pageSize, len(ahr.Txs))
	ahr.Txs = ahr.Txs[offsetStart:offsetEnd]

	return ahr, err
}
