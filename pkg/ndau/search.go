package ndau

import (
	metasearch "github.com/oneiro-ndev/metanode/pkg/meta/search"
	metatx "github.com/oneiro-ndev/metanode/pkg/meta/transaction"
)

// Implements AppSearchClient.
type NdauSearchClient struct {
	*metasearch.SearchClient

	maxHeight uint64
}

// Factory method.  Must call SearchClient.Init() before using the returned search client.
func NewNdauSearchClient() *NdauSearchClient {
	search := &NdauSearchClient{}
	search.SearchClient = metasearch.NewSearchClient()
	search.maxHeight = 0
	return search
}

// Reset our local cache for incrementally indexing the block at the given height.
func (search *NdauSearchClient) OnBeginBlock(height uint64) error {
	search.maxHeight = height // Only one block to consider for incremental indexing.
	return nil
}

// Grab the fields out of this transaction to index when the block is committed.
func (search *NdauSearchClient) OnDeliverTx(tx metatx.Transactable) error {
	// TODO: Implement.
	return nil
}

// Index all the transaction data we collected since the last BeginBlock().
func (search *NdauSearchClient) OnCommit() error {
	// TODO: Implement.
	return nil
}
