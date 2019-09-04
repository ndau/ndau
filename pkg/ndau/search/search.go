package search

// The public API for searching the index.

import (
	"encoding/base64"
	"sort"
	"strconv"

	"github.com/oneiro-ndev/ndau/pkg/query"
)

// StartTxHash is used by SearchTxTypes() to return the latest transactions on the blockchain.
const StartTxHash = "start"

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

// SearchTxTypes returns tx data for a range of transactions on or before the given tx hash.
func (search *Client) SearchTxTypes(txHash string, txTypes []string, limit int) (TxListValueData, error) {
	listValueData := TxListValueData{}

	if len(txTypes) == 0 {
		// No types given means no results.
		return listValueData, nil
	}

	var searchKeys []string
	for _, t := range txTypes {
		searchKeys = append(searchKeys, formatTxTypeToHeightSearchKey(t))
	}

	// Use a unique key name for each query.
	// NOTE: We could consider using a well-defined name based on the query params, but the
	// usefulness of that is primarily around multiple blockchain explorer clients hitting the
	// same endpoint near-simultaneously.  The most common one being the first page of results.
	// However, once we have higher volume, the latest page will be ever-changing.  And, in a low
	// volume ecosystem, there are likely not going to be many simultaneous queries.  So current
	// implementation uses a short-lived union key and we DELETE it, rather than EXPIRE it.
	searchKey := formatUniqueUnionSearchKey()

	// Delete the short-lived union key from the database when we're all done.
	defer search.Client.Del(searchKey)

	// Merge sort all the requested tx type results.
	_, err := search.Client.ZUnionStore(searchKey, searchKeys)
	if err != nil {
		return listValueData, err
	}

	// Get the rank of the starting tx hash.  If it's not in the list, we have a bad query.
	// Use reverse rank since we want transactions in reverse chronological order.
	var start int64
	if txHash == StartTxHash {
		// Start from the latest tx.
		start = 0

		// Exit early if there are no results.
		count, err := search.Client.ZCount(searchKey)
		if err != nil || count <= 0 {
			return listValueData, err
		}
	} else {
		start, err = search.Client.ZRevRank(searchKey, txHash)
		if err != nil {
			// No error if the hash is bad (not part of the results, invalid, etc).
			// Just return empty results.
			return listValueData, nil
		}
	}

	var stop int64
	if limit <= 0 {
		// Zero or negative limit means "unlimited".
		stop = -1
	} else {
		// Like the start rank, the stop rank is inclusive.
		// We don't subtract one here since we want to get the next tx hash after the page.
		// If this is more than there are results, all results after the start are returned.
		stop = start + int64(limit)
	}

	// Get a page's worth of results from the union.
	hashes, err := search.Client.ZRevRange(searchKey, start, stop)
	if err != nil {
		return listValueData, err
	}

	if limit <= 0 || limit >= len(hashes) {
		// We're getting the last page of results, there will be no "next" tx hash.
		limit = len(hashes)
	} else {
		// The last element is used for the next tx hash and not included in our results below.
		limit = len(hashes) - 1
		listValueData.NextTxHash = hashes[limit]
	}

	// Pull out transaction data using the tx hash index, for each tx hash in the list.
	for i := 0; i < limit; i++ {
		valueData, err := search.SearchTxHash(hashes[i])
		if err != nil {
			return listValueData, err
		}

		// Sanity check.  Every search in this loop is exepected to return a valid tx result.
		if valueData.BlockHeight > 0 {
			listValueData.Txs = append(listValueData.Txs, valueData)
		}
	}

	return listValueData, nil
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
