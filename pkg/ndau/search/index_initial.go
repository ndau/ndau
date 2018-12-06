package search

// The public API for initial indexing of the blockchain.

import (
	"fmt"

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
	search.state = nil
	search.txs = nil
	search.blockHash = ""
	search.blockHeight = 0
	search.nextHeight = 0

	// The height encountered on the previous iteration of the loop below.
	lastHeight := search.nextHeight

	// One more than the height we indexed to the last time we indexed the blockchain.
	// In other words, it's the height we want to index to this time.
	minHeightToIndex := search.Client.GetNextHeight()

	example := backing.State{}
	err = state.IterHistory(db, ds, &example, func(stI state.State, height uint64) error {
		// Save off the max height that we'll index up to by the end of the iteration.
		if search.nextHeight == 0 {
			// This assumes we're iterating blocks in order from the head to genesis.
			search.nextHeight = height + 1
			lastHeight = search.nextHeight
		}

		// If we've reached the last height we indexed to, we can stop here.
		if height < minHeightToIndex {
			// This assumes we're iterating blocks in order from the head to genesis.
			return state.StopIteration()
		}

		// Make sure we're iterating from the head block to genesis.
		// However, we support multiple height-0 entries for parallelism with chaos noms data.
		// And also at height 1 for unit tests.
		if height > lastHeight || height == lastHeight && height > 1 {
			// Indexing logic relies on this, but more importantly, this indicates
			// a serious problem in the blockchain if the height increases as we
			// crawl the blockchain from the head to the genesis block.
			panic(fmt.Sprintf("Invalid height found in noms: %d >= %d", height, lastHeight))
		}
		lastHeight = height

		search.blockHeight = height

		// Here's where we index data out of noms.
		// We will get block and tx hashes to index from our external indexer application.
		// TODO: Add something here to index.  If we don't need to index anything out of noms,
		// then remove this entire function.

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
