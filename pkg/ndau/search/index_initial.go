package search

// The public API for initial indexing of the blockchain.

import (
	"github.com/attic-labs/noms/go/datas"
	"github.com/oneiro-ndev/ndau/pkg/ndau/backing"
	"github.com/oneiro-ndev/metanode/pkg/meta/state"
)

// IndexBlockchain fills the index with data from the blockchain,
// from the head block down to just before the last block we indexed.
func (search *Client) IndexBlockchain(
	db datas.Database, ds datas.Dataset,
) (updateCount int, insertCount int, err error) {
	updateCount = 0
	insertCount = 0

	// Start fresh.  It should already be zero'd out upon entry.
	search.txHashes = nil
	search.blockHash = ""
	search.blockHeight = 0
	search.maxHeight = 0

	// The height encountered on the previous iteration of the loop below.
	lastHeight := search.maxHeight

	// The max height we indexed to the last time we indexed the blockchain.
	maxIndexedHeight := search.Client.GetHeight()

	example := backing.State{}
	err = state.IterHistory(db, ds, &example, func(stI state.State, height uint64) error {
		// Save off the max height that we'll index up to by the end of the iteration.
		if search.maxHeight == 0 {
			// This assumes we're iterating blocks in order from the head to genesis.
			search.maxHeight = height
			lastHeight = height + 1
		}

		// If we've reached the last height we indexed to, we can stop here.
		if height <= maxIndexedHeight {
			// This assumes we're iterating blocks in order from the head to genesis.
			return state.StopIteration()
		}

		// Make sure we're iterating from the head block to genesis.
		if height >= lastHeight {
			// Indexing logic relies on this, but more importantly, this indicates
			// a serious problem in the blockchain if the height increases as we
			// crawl the blockchain from the head to the genesis block.
			panic("Invalid height found in ndau chain")
		}
		lastHeight = height

		// TODO: Get hashes.
		//search.txHashes.append(txHash)
		//search.blockHash = blockHash
		search.blockHeight = height

		updCount, insCount, err := search.index()
		updateCount += updCount
		insertCount += insCount
		return err
	})
	if err != nil && !state.IsStopIteration(err) {
		return updateCount, insertCount, err
	}

	search.onIndexingComplete()

	return updateCount, insertCount, nil
}
