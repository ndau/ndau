package ndau

import (
	"fmt"
	"math/rand"
	"os/exec"
	"testing"
	"time"

	"github.com/oneiro-ndev/metanode/pkg/meta/app/code"
	metatx "github.com/oneiro-ndev/metanode/pkg/meta/transaction"
	"github.com/oneiro-ndev/ndau/pkg/ndau/backing"
	"github.com/oneiro-ndev/ndau/pkg/ndau/search"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	"github.com/oneiro-ndev/ndaumath/pkg/signature"
	math "github.com/oneiro-ndev/ndaumath/pkg/types"
	generator "github.com/oneiro-ndev/system_vars/pkg/genesis.generator"
	sv "github.com/oneiro-ndev/system_vars/pkg/system_vars"

	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
)

const sysvarKeys = "sysvar private keys"

func initAppSysvarWithIndex(t *testing.T, indexAddr string, indexVersion int) (
	*App, generator.Associated,
) {
	app, assc := initAppWithIndex(t, indexAddr, indexVersion)
	app.InitChain(abci.RequestInitChain{})

	// Fetch the sysvar address system variable.
	sysvarAddr := address.Address{}
	err := app.System(sv.SetSysvarAddressName, &sysvarAddr)
	require.NoError(t, err)
	assc[sysvarKeys], err = MockSystemAccount(app, sysvarAddr)

	return app, assc
}

func TestSysvarHistoryIndex(t *testing.T) {
	// Start redis server on a non-default (test) port.  Using a non-standard port means we
	// probably won't conflict with any other redis server currently running.  It also means we
	// disable persistence by not loading the default redis.conf file.  That way we don't have to
	// .gitignore it, and we don't have to force-wipe the index from the last time we ran the
	// tests.  We get a fresh test every time.
	portn := 6378 + rand.Intn(1024)
	port := fmt.Sprint(portn)
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
	app, assc := initAppSysvarWithIndex(t, "localhost:"+port, 0)

	// Test data.
	sysvar := "sysvar"
	value := "value"
	height := uint64(123)
	valueBytes := []byte(value)

	// Test incremental indexing.
	t.Run("TestSysvarHistoryIncrementalIndexing", func(t *testing.T) {
		// Begin a new block.
		begin := abci.RequestBeginBlock{Header: abci.Header{
			Time:   time.Now(),
			Height: int64(height),
		}}
		app.BeginBlock(begin)

		// Deliver an update.
		t.Run("TestSysvarHistoryIndexingTxUpdate", func(t *testing.T) {
			privateKeys := assc[sysvarKeys].([]signature.PrivateKey)
			tx := NewSetSysvar(
				sysvar,
				valueBytes,
				uint64(1),
				privateKeys[0],
			)

			txBytes, err := metatx.Marshal(tx, TxIDs)
			require.NoError(t, err)

			resp := app.CheckTx(txBytes)
			if resp.Log != "" {
				t.Log(resp.Log)
			}
			require.Equal(t, code.OK, code.ReturnCode(resp.Code))

			txResponse := app.DeliverTx(txBytes)
			require.Equal(t, "", txResponse.Log)
		})

		// End the block.
		end := abci.RequestEndBlock{}
		app.EndBlock(end)

		// Commit the block.
		app.Commit()
	})

	// Test searching.
	t.Run("TestSysvarHistorySearching", func(t *testing.T) {
		// Search for the update transaction we indexed.
		t.Run("TestSysvarHistorySearchingTxUpdate", func(t *testing.T) {
			hkr, err := app.GetSearch().(*search.Client).SearchSysvarHistory(sysvar, 0, 0)
			require.NoError(t, err)

			// Should have one result for our test key value pair.
			require.Equal(t, 1, len(hkr.History))
			require.Equal(t, height, hkr.History[0].Height)
			require.Equal(t, valueBytes, hkr.History[0].Value)
		})
	})
}

func TestIndex(t *testing.T) {
	// Start redis server on a non-default (test) port.  Using a non-standard port means we
	// probably won't conflict with any other redis server currently running.  It also means we
	// disable persistence by not loading the default redis.conf file.  That way we don't have to
	// .gitignore it, and we don't have to force-wipe the index from the last time we ran the
	// tests.  We get a fresh test every time.
	portn := 6378 + rand.Intn(1024)
	port := fmt.Sprint(portn)
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
	txOffset := 0                                 // One transaction in the block.
	tmBlockHash := []byte("abcdefghijklmnopqrst") // 20 bytes
	blockHash := fmt.Sprintf("%x", tmBlockHash)   // 40 characters
	var txHash string
	blockTime := time.Now()

	search := app.GetSearch().(*search.Client)

	// Test initial indexing.
	t.Run("TestHashInitialIndexing", func(t *testing.T) {
		updateCount, insertCount, err := search.IndexBlockchain(app.GetDB(), app.GetDS())
		require.NoError(t, err)

		// Number of sysvars present in noms.
		state := app.GetState().(*backing.State)
		numSysvars := len(state.Sysvars)

		// There should be nothing indexed outside of sysvars.
		require.Equal(t, 0, updateCount)
		require.Equal(t, numSysvars, insertCount)
	})

	// Test incremental indexing.
	t.Run("TestHashIncrementalIndexing", func(t *testing.T) {
		// Begin a new block.
		begin := abci.RequestBeginBlock{
			Hash: tmBlockHash,
			Header: abci.Header{
				Time:   blockTime,
				Height: int64(height),
			},
		}
		app.BeginBlock(begin)

		// Deliver a tx.
		t.Run("TestTxHashIndexing", func(t *testing.T) {
			privateKeys := assc[rfeKeys].([]signature.PrivateKey)
			rfe := NewReleaseFromEndowment(
				targetAddress,
				math.Ndau(1),
				uint64(1),
				privateKeys[0],
			)

			// Get the tx hash so we an search on it later.
			txHash = metatx.Hash(rfe)

			rfeBytes, err := metatx.Marshal(rfe, TxIDs)
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
			vd, err := search.SearchTxHash(txHash)
			require.NoError(t, err)
			require.Equal(t, height, vd.BlockHeight)
			require.Equal(t, txOffset, vd.TxOffset)
		})

		t.Run("TestAccountSearching", func(t *testing.T) {
			ahr, err := search.SearchAccountHistory(targetAddress.String(), 0, 0)
			require.NoError(t, err)
			require.Equal(t, 1, len(ahr.Txs))
			valueData := ahr.Txs[0]
			require.Equal(t, height, valueData.BlockHeight)
			require.Equal(t, txOffset, valueData.TxOffset)
			require.Equal(t, math.Ndau(1), valueData.Balance)
		})

		t.Run("TestDateRangeSearching", func(t *testing.T) {
			timeString := blockTime.Format(time.RFC3339)
			firstHeight, lastHeight, err := search.SearchDateRange(timeString, timeString)
			require.NoError(t, err)
			// Expecting the block before the one we indexed since it's flooring to current day.
			require.Equal(t, height-1, firstHeight)
			// Expecting the block after the one we indexed since it's an exclusive upper bound.
			require.Equal(t, height+1, lastHeight)
		})
	})
}
