package ndau

// ----- ---- --- -- -
// Copyright 2019 Oneiro NA, Inc. All Rights Reserved.
//
// Licensed under the Apache License 2.0 (the "License").  You may not use
// this file except in compliance with the License.  You can obtain a copy
// in the file LICENSE in the source distribution or at
// https://www.apache.org/licenses/LICENSE-2.0.txt
// - -- --- ---- -----

import (
	"fmt"
	"math/rand"
	"os/exec"
	"testing"
	"time"

	"github.com/oneiro-ndev/metanode/pkg/meta/app/code"
	metast "github.com/oneiro-ndev/metanode/pkg/meta/state"
	metatx "github.com/oneiro-ndev/metanode/pkg/meta/transaction"
	"github.com/oneiro-ndev/ndau/pkg/ndau/backing"
	srch "github.com/oneiro-ndev/ndau/pkg/ndau/search"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	"github.com/oneiro-ndev/ndaumath/pkg/constants"
	"github.com/oneiro-ndev/ndaumath/pkg/pricecurve"
	"github.com/oneiro-ndev/ndaumath/pkg/signature"
	math "github.com/oneiro-ndev/ndaumath/pkg/types"
	sv "github.com/oneiro-ndev/system_vars/pkg/system_vars"
	"github.com/stretchr/testify/require"
)

func withRedis(t *testing.T, test func(port string)) {
	// Start redis server on a non-default (test) port.  Using a non-standard port means we
	// probably won't conflict with any other redis server currently running.  It also means we
	// disable persistence by not loading the default redis.conf file.  That way we don't have to
	// .gitignore it, and we don't have to force-wipe the index from the last time we ran the
	// tests.  We get a fresh test every time.
	portn := 7000 + rand.Intn(1024)
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

	// run the test
	test(port)
}

func TestSysvarHistoryIndex(t *testing.T) {
	withRedis(t, func(port string) {
		// Create the app and tx factory.
		app, assc := initAppRFEWithIndex(t, "localhost:"+port, 0)

		// Test data.
		sysvar := "sysvar"
		value := "value"
		valueBytes := []byte(value)
		height := uint64(123)

		// Test incremental indexing.
		t.Run("TestSysvarHistoryIncrementalIndexing", func(t *testing.T) {
			privateKeys := assc[sysvarKeys].([]signature.PrivateKey)
			tx := NewSetSysvar(
				sysvar,
				valueBytes,
				uint64(1),
				privateKeys[0],
			)
			resp, _ := deliverTxContext(t, app, tx, ddc(t).atHeight(height))
			require.Equal(t, code.OK, code.ReturnCode(resp.Code))
		})

		// Test searching.
		t.Run("TestSysvarHistorySearching", func(t *testing.T) {
			// Search for the update transaction we indexed.
			hkr, err := app.GetSearch().(*srch.Client).SearchSysvarHistory(sysvar, 0, 0)
			require.NoError(t, err)

			// Should have one result for our test key value pair.
			require.Equal(t, 1, len(hkr.History))
			require.Equal(t, height, hkr.History[0].Height)
			require.Equal(t, valueBytes, hkr.History[0].Value)
		})
	})
}

func TestIndex(t *testing.T) {
	withRedis(t, func(port string) {
		// Create the app and tx factory.
		app, assc := initAppRFEWithIndex(t, "localhost:"+port, 0)

		// Test data.
		sysvar := "sysvar"
		value := "value"
		valueBytes := []byte(value)
		height := uint64(123)
		txOffsetSSV := 0
		txOffsetRFE := 1
		tmBlockHash := []byte("abcdefghijklmnopqrst") // 20 bytes
		blockHash := fmt.Sprintf("%x", tmBlockHash)   // 40 characters
		var txHashSSV, txHashRFE string
		blockTime, err := math.TimestampFrom(time.Now())
		require.NoError(t, err)

		search := app.GetSearch().(*srch.Client)

		// Ensure Redis is empty.
		err = search.FlushDB()
		require.NoError(t, err)

		// Test initial indexing.
		t.Run("TestHashInitialIndexing", func(t *testing.T) {
			_, insertCount, err := search.IndexBlockchain(app.GetDB(), app.GetDS())
			require.NoError(t, err)

			// Number of sysvars present in noms.
			state := app.GetState().(*backing.State)
			numSysvars := len(state.Sysvars)

			// The sysvars should all be inserted
			require.GreaterOrEqual(t, insertCount, numSysvars)
		})

		// Deliver some transactions, which should trigger incremental indexing
		privateKeys := assc[sysvarKeys].([]signature.PrivateKey)
		ssv := NewSetSysvar(
			sysvar,
			valueBytes,
			uint64(1),
			privateKeys[0],
		)
		txHashSSV = metatx.Hash(ssv)

		privateKeys = assc[rfeKeys].([]signature.PrivateKey)
		rfe := NewReleaseFromEndowment(
			targetAddress,
			math.Ndau(1),
			uint64(2),
			privateKeys[0],
		)
		txHashRFE = metatx.Hash(rfe)

		resps, _ := deliverTxsContext(t, app, []metatx.Transactable{
			ssv, rfe,
		}, ddc(t).withHash(blockHash).atHeight(height))
		for _, resp := range resps {
			require.Equal(t, code.OK, code.ReturnCode(resp.Code))
		}

		// Test searching.
		t.Run("TestHashSearching", func(t *testing.T) {
			t.Run("TestBlockHashSearching", func(t *testing.T) {
				heightResult, err := search.SearchBlockHash(blockHash)
				require.NoError(t, err)
				require.Equal(t, height, heightResult)
			})

			t.Run("TestTxHashSearchingSSV", func(t *testing.T) {
				vd, err := search.SearchTxHash(txHashSSV)
				require.NoError(t, err)
				require.Equal(t, height, vd.BlockHeight)
				require.Equal(t, txOffsetSSV, vd.TxOffset)
			})

			t.Run("TestTxHashSearchingRFE", func(t *testing.T) {
				vd, err := search.SearchTxHash(txHashRFE)
				require.NoError(t, err)
				require.Equal(t, height, vd.BlockHeight)
				require.Equal(t, txOffsetRFE, vd.TxOffset)
			})

			t.Run("TestTxTypeSearching", func(t *testing.T) {
				txTypes := []string{"ReleaseFromEndowment", "SetSysvar"}

				// The first query will tell us that this hash is the start of the next page.
				txHashNext := txHashSSV

				// The first page should return the latest (RFE) transaction.
				vd, err := search.SearchTxTypes("", txTypes, 1)
				require.NoError(t, err)
				require.Equal(t, 1, len(vd.Txs))
				require.Equal(t, txHashNext, vd.NextTxHash)
				require.Equal(t, height, vd.Txs[0].BlockHeight)
				require.Equal(t, txOffsetRFE, vd.Txs[0].TxOffset)

				// The second page should return the first (SSV) transaction.
				vd, err = search.SearchTxTypes(txHashNext, txTypes, 1)
				require.NoError(t, err)
				require.Equal(t, 1, len(vd.Txs))
				require.Equal(t, "", vd.NextTxHash)
				require.Equal(t, height, vd.Txs[0].BlockHeight)
				require.Equal(t, txOffsetSSV, vd.Txs[0].TxOffset)
			})

			t.Run("TestTxTypeSearchingTransfer", func(t *testing.T) {
				txTypes := []string{"Transfer"}
				vd, err := search.SearchTxTypes("", txTypes, 1)
				require.NoError(t, err)
				require.Equal(t, 0, len(vd.Txs))
				require.Equal(t, "", vd.NextTxHash)
			})

			t.Run("TestTxTypeSearchingInvalid", func(t *testing.T) {
				txTypes := []string{"ReleaseFromEndowment"}
				vd, err := search.SearchTxTypes("invalid-hash", txTypes, 1)
				require.NoError(t, err)
				require.Equal(t, 0, len(vd.Txs))
				require.Equal(t, "", vd.NextTxHash)
			})

			t.Run("TestTxTypeSearchingEmpty", func(t *testing.T) {
				vd, err := search.SearchTxTypes("invalid-hash", []string{}, 1)
				require.NoError(t, err)
				require.Equal(t, 0, len(vd.Txs))
				require.Equal(t, "", vd.NextTxHash)
			})

			t.Run("TestTxTypeSearchingNil", func(t *testing.T) {
				vd, err := search.SearchTxTypes("invalid-hash", nil, 1)
				require.NoError(t, err)
				require.Equal(t, 0, len(vd.Txs))
				require.Equal(t, "", vd.NextTxHash)
			})

			t.Run("TestAccountSearching", func(t *testing.T) {
				ahr, err := search.SearchAccountHistory(targetAddress.String(), 0, 0)
				require.NoError(t, err)
				require.Equal(t, 1, len(ahr.Txs))
				valueData := ahr.Txs[0]
				require.Equal(t, height, valueData.BlockHeight)
				require.Equal(t, txOffsetRFE, valueData.TxOffset)
				require.Equal(t, math.Ndau(1), valueData.Balance)
			})

			t.Run("TestDateRangeSearching", func(t *testing.T) {
				timeString := blockTime.String()
				firstHeight, lastHeight, err := search.SearchDateRange(timeString, timeString)
				require.NoError(t, err)
				// Expecting the block before the one we indexed since it's flooring to current day.
				require.Equal(t, height-1, firstHeight)
				// Expecting the block after the one we indexed since it's an exclusive upper bound.
				require.Equal(t, height+1, lastHeight)
			})
		})

		t.Run("TestMostRecentRegisterNode", func(t *testing.T) {
			// precondition: this node has never been registered
			txData, err := search.SearchMostRecentRegisterNode(targetAddress.String())
			require.NoError(t, err)
			require.Nil(t, txData)

			// setup: ensure node can be registered
			modify(t, targetAddress.String(), app, func(acct *backing.AccountData) {
				acct.Balance = 1000 * constants.NapuPerNdau
				acct.ValidationKeys = []signature.PublicKey{transferPublic}
			})
			noderules, _ := getRulesAccount(t, app)
			err = app.UpdateStateImmediately(app.Stake(
				1000*constants.NapuPerNdau,
				targetAddress, noderules, noderules,
				nil,
			))
			require.NoError(t, err)

			// state change: register the node.
			height := uint64(1024)
			rn := NewRegisterNode(targetAddress, []byte{0xa0, 0x00, 0x88}, targetPublic, 1, transferPrivate)
			resp, _ := deliverTxContext(t, app, rn, ddc(t).atHeight(height))
			require.Equal(t, code.OK, code.ReturnCode(resp.Code))

			// postcondition: MRRN must correctly identify the block height
			txData, err = search.SearchMostRecentRegisterNode(targetAddress.String())
			require.NoError(t, err)
			require.NotNil(t, txData)
			require.Equal(t, height, txData.BlockHeight)

			t.Run("Reregister", func(t *testing.T) {
				// precondition: the node has been deregistered
				app.UpdateStateImmediately(func(stateI metast.State) (metast.State, error) {
					state := stateI.(*backing.State)
					delete(state.Nodes, targetAddress.String())
					return state, nil
				})

				// intermediate test: MRRN still reports the most recent register node
				// tx, even though it no longer applies
				txData, err = search.SearchMostRecentRegisterNode(targetAddress.String())
				require.NoError(t, err)
				require.NotNil(t, txData)
				require.Equal(t, height, txData.BlockHeight)

				// state change: register the node.
				height = 4096
				rn := NewRegisterNode(targetAddress, []byte{0xa0, 0x00, 0x88}, targetPublic, 2, transferPrivate)
				resp, _ := deliverTxContext(t, app, rn, ddc(t).atHeight(height))
				require.Equal(t, code.OK, code.ReturnCode(resp.Code))

				// postcondition: MRRN must correctly identify the block height
				txData, err = search.SearchMostRecentRegisterNode(targetAddress.String())
				require.NoError(t, err)
				require.NotNil(t, txData)
				require.Equal(t, uint64(height), txData.BlockHeight)
			})
		})

		t.Run("TestSearchBlockTime", func(t *testing.T) {
			spairs := []struct {
				height uint64
				time   string
			}{
				{1021, "2019-05-04T03:02:01Z"},
				{1022, "2019-05-04T03:05:00Z"},
				{1023, "2019-05-04T03:08:00Z"},
			}

			sparsed := make([]struct {
				height uint64
				time   math.Timestamp
			}, len(spairs))

			var err error
			for idx, spair := range spairs {
				sparsed[idx].height = spair.height
				sparsed[idx].time, err = math.ParseTimestamp(spair.time)
				require.NoError(t, err, "must parse fixed timestamps")
			}

			// precondition: the search does not know about any block times
			for _, pair := range sparsed {
				time, err := search.BlockTime(pair.height)
				require.NoError(t, err)
				require.Zero(t, time)
			}

			// state change: generate blocks at each time/height pair
			for idx, pair := range sparsed {
				rfe := NewReleaseFromEndowment(targetAddress, 1, uint64(100+idx), privateKeys...)
				resp, _ := deliverTxContext(t, app, rfe, ddc(t).at(pair.time).atHeight(pair.height))
				require.Equal(t, code.OK, code.ReturnCode(resp.Code))
			}

			// postcondition: the search knows about all appropriate block times
			for _, pair := range sparsed {
				time, err := search.BlockTime(pair.height)
				require.NoError(t, err)
				require.Equal(t, pair.time, time)
			}
		})

		t.Run("TestMarketPriceSearch", func(t *testing.T) {
			// setup: generate but do not yet send some RecordPrice txs
			rpAddr := address.Address{}
			err := app.System(sv.RecordPriceAddressName, &rpAddr)
			require.NoError(t, err)
			rpPrivate, err := MockSystemAccount(app, rpAddr)
			require.NoError(t, err)

			type pair struct {
				ts math.Timestamp
				tx *RecordPrice
			}

			pairs := []pair{
				{1*math.Year + 1*math.Month + 1*math.Day, NewRecordPrice(5*pricecurve.Dollar, 1, rpPrivate...)},
				{1*math.Year + 5*math.Month + 16*math.Day, NewRecordPrice(12*pricecurve.Dollar, 2, rpPrivate...)},
			}

			// precondition: search does not know about any market price data
			priceResult, err := search.SearchMarketPrice(srch.PriceQueryParams{})
			require.NoError(t, err)
			require.Empty(t, priceResult.Items)
			require.False(t, priceResult.More, "must not have unreturned items")

			// state change: deliver the RecordPrice txs
			for idx, pair := range pairs {
				deliverTxContext(t, app, pair.tx, ddc(t).at(pair.ts).atHeight(23+(uint64(idx)*15)))
			}

			// postcondition: search can find the market price data
			t.Run("Unlimited", func(t *testing.T) {
				t.Parallel()
				priceResult, err := search.SearchMarketPrice(srch.PriceQueryParams{})
				require.NoError(t, err)
				require.Equal(t, len(pairs), len(priceResult.Items))
				require.False(t, priceResult.More, "must not have unreturned items")
				for idx := range pairs {
					require.Equal(t, pairs[idx].ts, priceResult.Items[idx].Timestamp)
					require.Equal(t, pairs[idx].tx.MarketPrice, priceResult.Items[idx].Price)
					require.NotZero(t, priceResult.Items[idx].Height)
				}
				require.NotEqual(t, priceResult.Items[0].Height, priceResult.Items[1].Height)
			})
			t.Run("Limited", func(t *testing.T) {
				t.Parallel()
				priceResult, err := search.SearchMarketPrice(srch.PriceQueryParams{
					Limit: 1,
				})
				require.NoError(t, err)
				require.Equal(t, 1, len(priceResult.Items))
				require.True(t, priceResult.More, "must have unreturned items")
				idx := 0 // endpoint interates fwd through history
				require.Equal(t, pairs[idx].ts, priceResult.Items[idx].Timestamp)
				require.Equal(t, pairs[idx].tx.MarketPrice, priceResult.Items[idx].Price)
				// get subsequent results
				priceResult, err = search.SearchMarketPrice(srch.PriceQueryParams{
					After: srch.RangeEndpoint{Timestamp: priceResult.Items[idx].Timestamp},
				})
				require.NoError(t, err)
				require.Equal(t, 1, len(priceResult.Items))
				require.False(t, priceResult.More, "must not have unreturned items")
				require.Equal(t, pairs[1].ts, priceResult.Items[idx].Timestamp)
				require.Equal(t, pairs[1].tx.MarketPrice, priceResult.Items[idx].Price)
			})
		})

		t.Run("TestTargetPriceSearch", func(t *testing.T) {
			// setup: generate but do not yet send some txs
			rfeAddr := address.Address{}
			err := app.System(sv.ReleaseFromEndowmentAddressName, &rfeAddr)
			require.NoError(t, err)
			privateKeys = assc[rfeKeys].([]signature.PrivateKey)
			modify(t, rfeAddr.String(), app, func(ad *backing.AccountData) {
				ad.Sequence = 5
			})

			renavAddr := address.Address{}
			err = app.System(sv.RecordEndowmentNAVAddressName, &renavAddr)
			require.NoError(t, err)
			renavPvt, err := MockSystemAccount(app, renavAddr)
			require.NoError(t, err)

			rpAddr := address.Address{}
			err = app.System(sv.RecordPriceAddressName, &rpAddr)
			require.NoError(t, err)
			rpPrivate, err := MockSystemAccount(app, rpAddr)
			require.NoError(t, err)

			type pair struct {
				ts  math.Timestamp
				txs []metatx.Transactable
				tgt pricecurve.Nanocent
			}

			qty := math.Ndau(100 * constants.NapuPerNdau)
			pairs := []pair{
				{ts: 2*math.Year + 1*math.Month + 1*math.Day, txs: []metatx.Transactable{
					NewReleaseFromEndowment(targetAddress, qty, 10, privateKeys...),
					NewIssue(qty, 11, privateKeys...),
				}},
				{ts: 2*math.Year + 5*math.Month + 16*math.Day, txs: []metatx.Transactable{
					NewRecordEndowmentNAV(10*1000*pricecurve.Dollar, 12, renavPvt...),
				}},
				{ts: 2*math.Year + 10*math.Month + 23*math.Day, txs: []metatx.Transactable{
					NewRecordPrice(22*pricecurve.Dollar, 20, rpPrivate...),
				}},
			}

			search.Client.FlushDB()

			// precondition: search does not know about any target price data
			priceResult, err := search.SearchTargetPrice(srch.PriceQueryParams{})
			require.NoError(t, err)
			require.Empty(t, priceResult.Items)
			require.False(t, priceResult.More, "must not have unreturned items")

			// state change: deliver the RecordPrice txs
			for idx, pair := range pairs {
				resps, _ := deliverTxsContext(t, app, pair.txs, ddc(t).at(pair.ts).atHeight(230+(uint64(idx)*15)))
				for txidx, resp := range resps {
					require.Equal(t,
						code.OK, code.ReturnCode(resp.Code),
						fmt.Sprintf("idx: %d; txidx: %d", idx, txidx),
					)
				}
				pairs[idx].tgt = app.GetState().(*backing.State).TargetPrice
			}

			// postcondition: search can find the target price data
			t.Run("Unlimited", func(t *testing.T) {
				t.Parallel()
				priceResult, err := search.SearchTargetPrice(srch.PriceQueryParams{})
				require.NoError(t, err)
				require.Equal(t, len(pairs), len(priceResult.Items))
				require.False(t, priceResult.More, "must not have unreturned items")
				for idx := range pairs {
					require.Equal(t, pairs[idx].ts, priceResult.Items[idx].Timestamp)
					require.Equal(t, pairs[idx].tgt, priceResult.Items[idx].Price)
					require.NotZero(t, priceResult.Items[idx].Height)
					if idx > 0 {
						require.NotEqual(t, priceResult.Items[idx-1].Height, priceResult.Items[idx].Height)
					}
				}
			})
			t.Run("Limited", func(t *testing.T) {
				t.Parallel()
				priceResult, err := search.SearchTargetPrice(srch.PriceQueryParams{
					Limit: 1,
				})
				require.NoError(t, err)
				require.Equal(t, 1, len(priceResult.Items))
				require.True(t, priceResult.More, "must have unreturned items")
				idx := 0 // endpoint interates fwd through history
				require.Equal(t, pairs[idx].ts, priceResult.Items[idx].Timestamp)
				require.Equal(t, pairs[idx].tgt, priceResult.Items[idx].Price)
				// get subsequent results
				priceResult, err = search.SearchTargetPrice(srch.PriceQueryParams{
					After: srch.RangeEndpoint{Timestamp: priceResult.Items[idx].Timestamp},
					Limit: 1,
				})
				require.NoError(t, err)
				require.Equal(t, 1, len(priceResult.Items))
				require.True(t, priceResult.More, "must have unreturned items")
				require.Equal(t, pairs[1].ts, priceResult.Items[idx].Timestamp)
				require.Equal(t, pairs[1].tgt, priceResult.Items[idx].Price)
			})
		})
	})
}
