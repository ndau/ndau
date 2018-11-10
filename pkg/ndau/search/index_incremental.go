package search

// Methods used for incremental indexing.

import (
	metatx "github.com/oneiro-ndev/metanode/pkg/meta/transaction"
)

// OnBeginBlock resets our local cache for incrementally indexing the block at the given height.
func (search *Client) OnBeginBlock(hash string, height uint64) error {
	// There's only one block to consider for incremental indexing.
	search.blockHash = hash
	search.blockHeight = height
	search.maxHeight = height
	return nil
}

// OnDeliverTx grabs the fields out of this transaction to index when the block is committed.
func (search *Client) OnDeliverTx(tx metatx.Transactable) error {
	// We ignore all transactions.  We currently only care to index the block hash with height.
	return nil
}

// OnCommit indexes all the transaction data we collected since the last BeginBlock().
func (search *Client) OnCommit() error {
	_, _, err := search.onIndexingComplete()
	return err
}
