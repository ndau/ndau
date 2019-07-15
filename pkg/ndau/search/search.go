package search

// The public API for searching the index.

import (
	"encoding/base64"
	"sort"
	"strconv"

	"github.com/oneiro-ndev/ndau/pkg/query"
)

// SearchSysvarHistory returns value history for the given sysvar using an index under the hood.
// The response is sorted by ascending block height, each entry is where the key's value changed.
// Pass in 0,0 for the paging params to get the entire history.
func (search *Client) SearchSysvarHistory(
	sysvar string, afterHeight uint64, limit int,
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

		height := valueData.Height
		valueBase64 := valueData.ValueBase64

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

	// Reduce the full results list down to the requested portion.  There is some wasted effort with
	// this approach, but we support the worst case, which is to return all results.  In practice,
	// getting the full list from the underlying index is fast, with tolerable sorting speed.
	offsetStart := sort.Search(len(khr.History), func(n int) bool {
		return khr.History[n].Height > afterHeight
	})
	khr.History = khr.History[offsetStart:]
	if limit > 0 && len(khr.History) > limit {
		khr.History = khr.History[:limit]
	}

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

// SearchTxHash returns tx data for the given tx hash.
//
// Returns nil and no error if the given tx hash was not found in the index.
func (search *Client) SearchTxHash(txHash string) (TxValueData, error) {
	valueData := TxValueData{}
	searchKey := formatTxHashToHeightSearchKey(txHash)

	searchValue, err := search.Client.Get(searchKey)
	if err != nil {
		return valueData, err
	}

	if searchValue == "" {
		// Empty search results.  No error.
		return valueData, err
	}

	err = valueData.Unmarshal(searchValue)
	return valueData, err
}

// SearchAccountHistory returns an array of block height and txoffset pairs associated with the
// given account address.
// Pass in 0, 0 for the paging params to get the entire history.
func (search *Client) SearchAccountHistory(
	addr string, afterHeight uint64, limit int,
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

	// Reduce the full results list down to the requested portion.  There is some wasted effort with
	// this approach, but we support the worst case, which is to return all results.  In practice,
	// getting the full list from the underlying index is fast, with tolerable sorting speed.
	offsetStart := sort.Search(len(ahr.Txs), func(n int) bool {
		return ahr.Txs[n].BlockHeight > afterHeight
	})
	ahr.Txs = ahr.Txs[offsetStart:]
	if limit > 0 && len(ahr.Txs) > limit {
		ahr.Txs = ahr.Txs[:limit]
	}

	return ahr, err
}
