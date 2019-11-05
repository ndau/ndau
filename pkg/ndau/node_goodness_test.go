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
	"encoding/base64"
	"fmt"
	"sort"
	"testing"
	"time"

	"github.com/oneiro-ndev/metanode/pkg/meta/app/code"
	metast "github.com/oneiro-ndev/metanode/pkg/meta/state"
	metatx "github.com/oneiro-ndev/metanode/pkg/meta/transaction"
	"github.com/oneiro-ndev/msgp-well-known-types/wkt"
	"github.com/oneiro-ndev/ndau/pkg/ndau/backing"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	"github.com/oneiro-ndev/ndaumath/pkg/constants"
	"github.com/oneiro-ndev/ndaumath/pkg/signature"
	math "github.com/oneiro-ndev/ndaumath/pkg/types"
	sv "github.com/oneiro-ndev/system_vars/pkg/system_vars"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
)

func initAppNodeGoodness(t *testing.T, goodnesses ...uint64) (*App, []goodnessPair) {
	app, _ := initApp(t)
	gm := make(map[string]uint64)
	out := make([]goodnessPair, 0, len(goodnesses))
	rules, _ := getRulesAccount(t, app)
	const stakeQty = 1000 * constants.NapuPerNdau
	now, err := math.TimestampFrom(time.Now())
	require.NoError(t, err)

	for range goodnesses {
		pubkey, pvtkey, err := signature.Generate(signature.Ed25519, nil)
		require.NoError(t, err)
		addr, err := address.Generate(address.KindNdau, pubkey.KeyBytes())
		require.NoError(t, err)
		modify(t, addr.String(), app, func(ad *backing.AccountData) {
			ad.Balance = stakeQty
			ad.ValidationKeys = []signature.PublicKey{pubkey}
		})
		ensureRecent(t, app, addr.String())

		txs := []metatx.Transactable{
			NewStake(addr, rules, rules, stakeQty, 1, pvtkey),
			NewRegisterNode(addr, []byte{0xa0, 0x00, 0x88}, pubkey, 2, pvtkey),
		}
		resps, _ := deliverTxsContext(t, app, txs, ddc(t).at(now-math.Year))
		for _, resp := range resps {
			require.Equal(t, code.OK, code.ReturnCode(resp.Code))
		}

		out = append(out, goodnessPair{
			addr: addr.String(),
		})
	}
	// now sort the goodnesses alphabetically, so tests can always assume that
	// (all else being equal) they can predict the winner simply according to
	// the input goodnesses
	sort.Slice(out, func(i, j int) bool { return out[i].addr < out[j].addr })
	for i := 0; i < len(goodnesses); i++ {
		out[i].goodness = goodnesses[i]
		gm[out[i].addr] = out[i].goodness
	}
	app.goodnessFunc = func(addr string) (int64, error) {
		return int64(gm[addr]), nil
	}
	return app, out
}

func setMaxRewarded(t *testing.T, app *App, maxRewarded uint64) {
	app.UpdateStateImmediately(func(stI metast.State) (metast.State, error) {
		state := stI.(*backing.State)
		var err error
		state.Sysvars[sv.NodeRewardMaxRewarded], err = wkt.Uint64(maxRewarded).MarshalMsg(nil)
		require.NoError(t, err)
		return state, nil
	})
}

func TestNodeGoodness(t *testing.T) {
	app, gs := initAppNodeGoodness(t, 1, 1)
	winner, err := app.SelectByGoodness(0)
	require.NoError(t, err)
	require.Equal(t, gs[0].addr, winner.String())
}

func TestNodeGoodnessNoNodes(t *testing.T) {
	app, _ := initAppNodeGoodness(t)
	winner, err := app.SelectByGoodness(0)
	require.Error(t, err)
	require.Zero(t, winner)
}

func TestNodeGoodnessCanSelectSecond(t *testing.T) {
	app, gs := initAppNodeGoodness(t, 1, 1)
	random := (uint64(1) << 63) + 1
	winner, err := app.SelectByGoodness(random)
	require.NoError(t, err)
	require.Equal(t, gs[1].addr, winner.String())
}

func TestNodeGoodnessPicksFromBest(t *testing.T) {
	app, gs := initAppNodeGoodness(t, 1, 2, 4, 8, 16, 32)
	setMaxRewarded(t, app, 2)
	winner, err := app.SelectByGoodness(0)
	require.NoError(t, err)
	require.Equal(t, gs[4].addr, winner.String())
}

func TestNodeGoodnessExpandsTiesAppropriately(t *testing.T) {
	app, gs := initAppNodeGoodness(t, 1, 2, 2)
	setMaxRewarded(t, app, 1)
	winner, err := app.SelectByGoodness(0)
	require.NoError(t, err)
	require.Equal(t, gs[1].addr, winner.String())
	winner, err = app.SelectByGoodness(^uint64(0))
	require.NoError(t, err)
	require.Equal(t, gs[2].addr, winner.String())
}

func TestNodeGoodnessAndValidatorSets(t *testing.T) {
	type gst struct {
		name string
		gs   []uint64
		eo   int
	}

	cases := []gst{
		{name: "empty", gs: []uint64{}},
		{name: "one", gs: []uint64{1}},
		{name: "two", gs: []uint64{1, 1}, eo: 2},
		{name: "powers", gs: []uint64{1, 2, 4, 8, 16, 32}},
		{name: "revpowers", gs: []uint64{32, 16, 8, 4, 2, 1}},
	}

	t.Run("BeforeSysvar", func(t *testing.T) {
		for _, tcase := range cases {
			t.Run(tcase.name, func(t *testing.T) {
				app, _ := initAppNodeGoodness(t, tcase.gs...)
				reb := app.EndBlock(abci.RequestEndBlock{})
				require.Empty(t, reb.ValidatorUpdates)
			})
		}
	})
	t.Run("WithSysvar", func(t *testing.T) {
		for _, tcase := range cases {
			t.Run(tcase.name, func(t *testing.T) {
				for _, svValue := range []uint64{1, 2, 4} {
					t.Run(fmt.Sprint(svValue), func(t *testing.T) {
						expect := int(svValue)
						if tcase.eo > 0 {
							expect = tcase.eo
						}
						if len(tcase.gs) < expect {
							expect = len(tcase.gs)
						}

						app, _ := initAppNodeGoodness(t, tcase.gs...)
						app.UpdateStateImmediately(func(stI metast.State) (metast.State, error) {
							state := stI.(*backing.State)
							var err error
							state.Sysvars[sv.NodeMaxValidators], err = wkt.Uint64(svValue).MarshalMsg(nil)
							require.NoError(t, err)
							return state, nil
						})

						reb := app.EndBlock(abci.RequestEndBlock{})
						require.Equal(t, expect, len(reb.ValidatorUpdates))
					})
				}
			})
		}
	})
}

func TestNodeGoodnesses(t *testing.T) {
	app, gs := initAppNodeGoodness(t, 1, 1)
	// we use the real node goodness function here, meaning that we can't predict
	// the value of the node goodnesses, just that we will have the correct qty
	app.goodnessFunc = app.goodnessOf
	// set the real goodness function
	script := "oAAsDgZBJgAQpdToACwrtXy1TJMrAgBBRgUmABCl1OgARgUmABCl1OgARg8FGkIOA0AOA4IAe4IBfAVwe5AJcHyQQAUhFCECQsSKIBCPIRQhAkImABCl1OgACUZJJgAQpdToAEAmABCl1OgARg4DJgAQpdToAEaIgAAAYHmJIQKOII+IgAEAYHqKIQWOII+I"
	scriptB, err := base64.StdEncoding.DecodeString(script)
	require.NoError(t, err)
	app.UpdateStateImmediately(func(stI metast.State) (metast.State, error) {
		state := stI.(*backing.State)
		var err error
		state.Sysvars[sv.NodeGoodnessFuncName], err = wkt.Bytes(scriptB).MarshalMsg(nil)
		require.NoError(t, err)
		return state, nil
	})
	// begin block to set the current block time
	app.BeginBlock(abci.RequestBeginBlock{
		Header: abci.Header{
			Time: time.Now(),
		},
	})

	gs, sum := nodeGoodnesses(app)
	require.NotEmpty(t, gs)
	require.NotZero(t, sum)

	t.Run("topn", func(t *testing.T) {
		gs = topNGoodnesses(gs, 5)
		require.NotEmpty(t, gs)
	})
}
