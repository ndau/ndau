package ndau

import (
	"sort"
	"testing"

	"github.com/oneiro-ndev/metanode/pkg/meta/app/code"
	metast "github.com/oneiro-ndev/metanode/pkg/meta/state"
	"github.com/oneiro-ndev/msgp-well-known-types/wkt"
	"github.com/oneiro-ndev/ndau/pkg/ndau/backing"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	"github.com/oneiro-ndev/ndaumath/pkg/constants"
	"github.com/oneiro-ndev/ndaumath/pkg/signature"
	sv "github.com/oneiro-ndev/system_vars/pkg/system_vars"
	"github.com/stretchr/testify/require"
)

func initAppNodeGoodness(t *testing.T, goodnesses ...uint64) (*App, []goodnessPair) {
	app, _ := initApp(t)
	gm := make(map[string]uint64)
	out := make([]goodnessPair, 0, len(goodnesses))
	rules, _ := getRulesAccount(t, app)
	const stakeQty = 1000 * constants.NapuPerNdau
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

		var tx NTransactable
		tx = NewStake(addr, rules, rules, stakeQty, 1, pvtkey)
		resp := deliverTx(t, app, tx)
		require.Equal(t, code.OK, code.ReturnCode(resp.Code))
		tx = NewRegisterNode(addr, []byte{0xa0, 0x00, 0x88}, pubkey, 2, pvtkey)
		resp = deliverTx(t, app, tx)
		require.Equal(t, code.OK, code.ReturnCode(resp.Code))

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
