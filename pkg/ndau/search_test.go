package ndau

import (
	"fmt"
	"os/exec"
	"testing"
	"time"

	"github.com/oneiro-ndev/metanode/pkg/meta/app/code"
	metatx "github.com/oneiro-ndev/metanode/pkg/meta/transaction"
	"github.com/oneiro-ndev/ndau/pkg/ndau/search"
	"github.com/oneiro-ndev/ndaumath/pkg/signature"
	math "github.com/oneiro-ndev/ndaumath/pkg/types"

	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
)

func TestHashIndex(t *testing.T) {
	// Start redis server on a non-default (test) port.  Using a non-standard port means we
	// probably won't conflict with any other redis server currently running.  It also means we
	// disable persistence by not loading the default redis.conf file.  That way we don't have to
	// .gitignore it, and we don't have to force-wipe the index from the last time we ran the
	// tests.  We get a fresh test every time.
	port := "6378"
	cmd := exec.Command("redis-server", "--port", port)
	cmd.Start()

	// If redis isn't installed, we'll skip the test.
	if cmd.Process == nil {
		t.Skip("Unable to launch redis server; is it installed and in your $PATH?")
	}

	// Kill this process when the test exits, success or failure.
	defer cmd.Process.Kill()

	// Give it some time to ready itself.
	time.Sleep(time.Second)

	// Create the app and tx factory.
	app, assc := initAppRFEWithIndex(t, "localhost:"+port, 0)

	// Test data.
	height := uint64(123)
	txOffset := 0 // One transaction in the block.
	tmBlockHash := []byte("abcdefghijklmnopqrst") // 20 bytes
	blockHash := fmt.Sprintf("%x", tmBlockHash) // 40 characters
	var txHash string

	search := app.GetSearch().(*search.Client)

	// Test initial indexing.
	t.Run("TestHashInitialIndexing", func(t *testing.T) {
		updateCount, insertCount, err := search.IndexBlockchain(app.GetDB(), app.GetDS())
		require.NoError(t, err)

		// There should be nothing indexed.
		require.Equal(t, 0, updateCount)
		require.Equal(t, 0, insertCount)
	})

	// Test incremental indexing.
	t.Run("TestHashIncrementalIndexing", func(t *testing.T) {
		// Begin a new block.
		begin := abci.RequestBeginBlock{
			Hash: tmBlockHash,
			Header: abci.Header{
				Time:   time.Now(),
				Height: int64(height),
			},
		}
		app.BeginBlock(begin)

		// Deliver a tx.
		t.Run("TestTxHashIndexing", func(t *testing.T) {
			privateKeys := assc[rfeKeys].([]signature.PrivateKey)
			rfe := NewReleaseFromEndowment(
				math.Ndau(1),
				targetAddress,
				uint64(1),
				[]signature.PrivateKey{privateKeys[0]},
			)

			// Get the tx hash so we an search on it later.
			txHash = metatx.Hash(&rfe)

			rfeBytes, err := metatx.Marshal(&rfe, TxIDs)
			require.NoError(t, err)

			resp := app.CheckTx(rfeBytes)
			if resp.Log != "" {
				t.Log(resp.Log)
			}
			require.Equal(t, code.OK, code.ReturnCode(resp.Code))

			txResponse := app.DeliverTx(rfeBytes)
			require.Equal(t, "", txResponse.Log)
		})

		// End the block.
		end := abci.RequestEndBlock{}
		app.EndBlock(end)

		// Commit the block.
		app.Commit()
	})

	// Test searching.
	t.Run("TestHashSearching", func(t *testing.T) {
		t.Run("TestBlockHashSearching", func(t *testing.T) {
			heightResult, err := search.SearchBlockHash(blockHash)
			require.NoError(t, err)
			require.Equal(t, height, heightResult)
		})

		t.Run("TestTxHashSearching", func(t *testing.T) {
			heightResult, txOffsetResult, err := search.SearchTxHash(txHash)
			require.NoError(t, err)
			require.Equal(t, height, heightResult)
			require.Equal(t, txOffset, txOffsetResult)
		})
	})
}
