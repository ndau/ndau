package ndau

import (
	"testing"
	"time"

	"github.com/oneiro-ndev/metanode/pkg/meta/transaction"

	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/abci/types"
	"golang.org/x/crypto/ed25519"
)

func validatorOptCk(t *testing.T, power int64, app *App, check bool) (gtvc GTValidatorChange) {
	public, _, err := ed25519.GenerateKey(nil)
	require.NoError(t, err)

	gtvc = GTValidatorChange{
		PublicKey: public,
		Power:     power,
	}
	tx, err := metatx.Marshal(&gtvc, TxIDs)
	require.NoError(t, err)

	if check {
		result := app.CheckTx(tx)
		require.False(t, result.IsErr(), result.String())
	}

	return
}

func validator(t *testing.T, power int64, app *App) GTValidatorChange {
	return validatorOptCk(t, power, app, true)
}

func toVals(gtvcs []GTValidatorChange) (vals []types.ValidatorUpdate) {
	for _, gtvc := range gtvcs {
		vals = append(vals, gtvc.ToValidator())
	}
	return
}

func updateValidators(t *testing.T, app *App, updates []GTValidatorChange) {
	app.BeginBlock(types.RequestBeginBlock{Header: types.Header{
		Time: time.Now(),
	}})
	for _, gtvc := range updates {
		tx, err := metatx.Marshal(&gtvc, TxIDs)
		require.NoError(t, err)

		response := app.DeliverTx(tx)
		require.True(t, response.IsOK())
	}

	ebResp := app.EndBlock(types.RequestEndBlock{})
	actual := ebResp.GetValidatorUpdates()
	expect := make([]types.ValidatorUpdate, 0, len(updates))
	for _, gtvc := range updates {
		expect = append(expect, gtvc.ToValidator())
	}
	t.Logf("expect: %q", expect)
	t.Logf("actual: %q", actual)
	require.ElementsMatch(t, expect, ebResp.GetValidatorUpdates())

	app.Commit()
}

func initAppValidators(t *testing.T, valQty int) (app *App, gtvcs []GTValidatorChange) {
	app, _ = initApp(t)

	gtvcs = make([]GTValidatorChange, 0, valQty)
	validators := make([]types.ValidatorUpdate, 0, valQty)

	for i := 0; i < valQty; i++ {
		gtvc := validatorOptCk(t, 1, app, false)
		gtvcs = append(gtvcs, gtvc)
		validators = append(validators, gtvc.ToValidator())
	}

	// init the chain with those validators
	app.InitChain(types.RequestInitChain{Validators: validators})

	return
}

// Test basic operations on GTVC transactions
func TestGTValidatorChange(t *testing.T) {
	initAppValidators(t, 1)
}
