package search

// Methods used for incremental indexing.

import (
	"encoding/base64"
	"time"

	meta "github.com/oneiro-ndev/metanode/pkg/meta/app"
	metatx "github.com/oneiro-ndev/metanode/pkg/meta/transaction"
)

// OnBeginBlock resets our local cache for incrementally indexing the block at the given height.
func (search *Client) OnBeginBlock(height uint64, blockTime time.Time, tmHash string) error {
	// There's only one block to consider for incremental indexing.
	search.sysvarKeyToValueData = make(map[string]*ValueData)
	search.state = nil
	search.txs = nil
	search.blockTime = blockTime
	search.blockHash = tmHash
	search.blockHeight = height
	search.nextHeight = height + 1
	return nil
}

// OnDeliverTx grabs the fields out of this transaction to index when the block is committed.
func (search *Client) OnDeliverTx(tx metatx.Transactable) error {
	search.txs = append(search.txs, tx)

	indexable, ok := tx.(SysvarIndexable)
	if !ok {
		// This transactable is not set up to be indexable, perform a successful no-op.
		return nil
	}

	key := indexable.GetName()
	valueBase64 := base64.StdEncoding.EncodeToString(indexable.GetValue())

	searchKey := formatSysvarKeyToValueSearchKey(key)
	data, hasValue := search.sysvarKeyToValueData[searchKey]
	if hasValue {
		// Override whatever value was there before for this block.
		// We only want one k-v pair per block height in our index: the one for the latest value.
		data.valueBase64 = valueBase64
	} else {
		search.sysvarKeyToValueData[searchKey] = &ValueData{
			height:      search.blockHeight,
			valueBase64: valueBase64,
		}
	}

	return nil
}

// OnCommit indexes all the transaction data we collected since the last BeginBlock().
func (search *Client) OnCommit(app *meta.App) error {
	search.state = app.GetState()

	_, _, err := search.index()
	if err != nil {
		return err
	}

	// We don't need to check for dupes in the new block since we filtered them out by using the
	// sysvarKeyToValueData map.  However we do need to check for dupes from values in earlier
	// blocks.
	_, _, err = search.onIndexingComplete(true)

	return err
}
