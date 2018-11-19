package search

// Methods used for incremental indexing.

import (
	metatx "github.com/oneiro-ndev/metanode/pkg/meta/transaction"
)

// OnBeginBlock resets our local cache for incrementally indexing the block at the given height.
func (search *Client) OnBeginBlock(height uint64, tmHash string) error {
	// There's only one block to consider for incremental indexing.
	search.txHashes = nil
	search.blockHash = tmHash
	search.blockHeight = height
	search.maxHeight = height
	return nil
}

// OnDeliverTx grabs the fields out of this transaction to index when the block is committed.
func (search *Client) OnDeliverTx(tx metatx.Transactable) error {
	search.txHashes = append(search.txHashes, metatx.Hash(tx))
	return nil
}

// OnCommit indexes all the transaction data we collected since the last BeginBlock().
func (search *Client) OnCommit(appHash string) error {
	search.index()
	search.onIndexingComplete()
	return nil
}
