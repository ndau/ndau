package ndau

import (
	"fmt"
	"testing"

	generator "github.com/oneiro-ndev/chaos_genesis/pkg/genesis.generator"
	"github.com/oneiro-ndev/metanode/pkg/meta/app/code"
	metast "github.com/oneiro-ndev/metanode/pkg/meta/state"
	metatx "github.com/oneiro-ndev/metanode/pkg/meta/transaction"
	"github.com/oneiro-ndev/ndau/pkg/ndau/backing"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	"github.com/oneiro-ndev/ndaumath/pkg/signature"
	sv "github.com/oneiro-ndev/system_vars/pkg/system_vars"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/abci/types"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto/ed25519"
)

const cvcKeys = "cvc private keys"

func initAppCVC(t *testing.T) (*App, generator.Associated) {
	app, assc := initApp(t)
	app.InitChain(abci.RequestInitChain{})

	// fetch the CVC address system variable
	cvcAddr := address.Address{}
	err := app.System(sv.CommandValidatorChangeAddressName, &cvcAddr)
	require.NoError(t, err)
	assc[cvcKeys], err = MockSystemAccount(app, cvcAddr)

	return app, assc
}

// construct a new tendermint representation of an ed25519 key
func makepub() []byte {
	pk := ed25519.GenPrivKey().PubKey().(ed25519.PubKeyEd25519)
	return []byte(pk[:])
}
func TestCVCIsValidWithValidSignature(t *testing.T) {
	app, assc := initAppCVC(t)
	privateKeys := assc[cvcKeys].([]signature.PrivateKey)

	for i := 0; i < len(privateKeys); i++ {
		private := privateKeys[i]
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			cvc := NewCommandValidatorChange(
				makepub(),
				1,
				1,
				private,
			)

			cvcBytes, err := metatx.Marshal(cvc, TxIDs)
			require.NoError(t, err)

			resp := app.CheckTx(cvcBytes)
			if resp.Log != "" {
				t.Log(resp.Log)
			}
			require.Equal(t, code.OK, code.ReturnCode(resp.Code))
		})
	}
}

func TestCVCPublicKeyMustNotBeEmpty(t *testing.T) {
	app, assc := initAppCVC(t)
	privateKeys := assc[cvcKeys].([]signature.PrivateKey)
	private := privateKeys[0]

	for _, zv := range []struct {
		name string
		zero []byte
	}{
		{"nil", nil},
		{"empty", []byte{}},
	} {
		t.Run(zv.name, func(t *testing.T) {
			cvc := NewCommandValidatorChange(
				zv.zero,
				1,
				1,
				private,
			)

			cvcBytes, err := metatx.Marshal(cvc, TxIDs)
			require.NoError(t, err)

			resp := app.CheckTx(cvcBytes)
			t.Log(resp.Log)
			require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
		})
	}
}

func TestCVCPublicKeyMustBeCorrectSize(t *testing.T) {
	app, assc := initAppCVC(t)
	privateKeys := assc[cvcKeys].([]signature.PrivateKey)
	private := privateKeys[0]

	// public keys must be exactly 32 bytes long
	for i := 30; i < 35; i++ {
		if i != ed25519.PubKeyEd25519Size {
			t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
				cvc := NewCommandValidatorChange(
					make([]byte, i),
					1,
					1,
					private,
				)

				cvcBytes, err := metatx.Marshal(cvc, TxIDs)
				require.NoError(t, err)

				resp := app.CheckTx(cvcBytes)
				t.Log(resp.Log)
				require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
			})
		}
	}
}

func TestCVCIsInvalidWithInvalidSignature(t *testing.T) {
	app, _ := initAppCVC(t)
	_, private, err := signature.Generate(signature.Ed25519, nil)

	cvc := NewCommandValidatorChange(
		makepub(),
		1,
		1,
		private,
	)

	cvcBytes, err := metatx.Marshal(cvc, TxIDs)
	require.NoError(t, err)

	resp := app.CheckTx(cvcBytes)
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
}

func TestCVCIsValidOnlyWithSufficientTxFee(t *testing.T) {
	app, assc := initAppCVC(t)
	privateKeys := assc[cvcKeys].([]signature.PrivateKey)

	txFeeAddr := address.Address{}
	err := app.System(sv.CommandValidatorChangeAddressName, &txFeeAddr)
	require.NoError(t, err)

	// with a tx fee of 1, only the first tx should succeed
	modify(t, txFeeAddr.String(), app, func(ad *backing.AccountData) {
		ad.Balance = 1
	})

	// our fixtures are set up with 2 cvc keys
	for i := 0; i < len(privateKeys); i++ {
		private := privateKeys[i]
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			cvc := NewCommandValidatorChange(
				makepub(),
				1,
				uint64(i)+1,
				private,
			)

			resp := deliverTxWithTxFee(t, app, cvc)

			var expect code.ReturnCode
			if i == 0 {
				expect = code.OK
			} else {
				expect = code.InvalidTransaction
			}
			require.Equal(t, expect, code.ReturnCode(resp.Code))
		})
	}
}

// we can't really test that CVC transactions actually update the validator
// set within a unit test: the actual validator set, the one which matters,
// is stored in tendermint. To actually test this behavior, we need to write
// integration tests.
//
// Here's what we _can_ do: we can write tests which ensure that when a
// CVC transactable is sent up, the metanode does the right thing, and
// sends the right update list along with the relevant EndBlock response.
//
// Those tests live below, along with some test helpers to make them work.

// convert a list of cvcs into a list of validators
func toVals(cvcs []CommandValidatorChange) (vals []abci.ValidatorUpdate) {
	for _, cvc := range cvcs {
		vals = append(vals, cvc.ToValidator())
	}
	return
}

// send every update in the list of validator changes to the metanode,
// and ensure that the metanode has kept track of it and returns it in
// the EndBlock transaction
func updateValidators(t *testing.T, app *App, updates []CommandValidatorChange) {
	metatxs := make([]metatx.Transactable, len(updates))
	for i := 0; i < len(updates); i++ {
		metatxs[i] = metatx.Transactable(&updates[i])
	}

	resps, ebResp := deliverTxsContext(t, app, metatxs, ddc(t))
	for _, resp := range resps {
		require.Equal(t, code.OK, code.ReturnCode(resp.Code))
	}

	actual := ebResp.GetValidatorUpdates()
	expect := make([]types.ValidatorUpdate, len(updates))
	for i := 0; i < len(updates); i++ {
		expect[i] = updates[i].ToValidator()
	}

	t.Logf("expect: %q", expect)
	t.Logf("actual: %q", actual)
	require.ElementsMatch(t, expect, ebResp.GetValidatorUpdates())

	app.Commit()
}

func initAppCVCValidators(
	t *testing.T,
	valQty int,
) (app *App, ma generator.Associated, vcs []CommandValidatorChange) {
	app, ma = initAppCVC(t)

	vcs = make([]CommandValidatorChange, 0, valQty)
	validators := make([]abci.ValidatorUpdate, 0, valQty)

	for i := 0; i < valQty; i++ {
		cvc := NewCommandValidatorChange(
			makepub(),
			1,
			uint64(i)+1,
			ma[cvcKeys].([]signature.PrivateKey)...,
		)
		vcs = append(vcs, *cvc)
		validators = append(validators, cvc.ToValidator())
	}

	// set up these validators as the initial ones on the chain
	app.InitChain(abci.RequestInitChain{Validators: validators})

	return
}

func TestCommandValidatorChangeInit(t *testing.T) {
	initAppCVCValidators(t, 1)
}

func TestCommandValidatorChangeInitChain(t *testing.T) {
	qtyVals := 10
	app, _, cvcs := initAppCVCValidators(t, qtyVals)

	actualValidators, err := app.Validators()
	require.NoError(t, err)
	metast.ValidatorsAreEquivalent(
		t,
		metast.ValUpdatesToVals(t, toVals(cvcs)),
		actualValidators,
	)
}

func TestCommandValidatorChangeAddValidator(t *testing.T) {
	const qtyVals = 1
	app, ma, cvcs := initAppCVCValidators(t, qtyVals)

	// add a validator
	newCVC := NewCommandValidatorChange(
		makepub(),
		1,
		qtyVals+1,
		ma[cvcKeys].([]signature.PrivateKey)...,
	)
	require.NotNil(t, newCVC)
	cvcs = append(cvcs, *newCVC)
	updateValidators(t, app, []CommandValidatorChange{*newCVC})

	actualValidators, err := app.Validators()
	require.NoError(t, err)
	metast.ValidatorsAreEquivalent(
		t,
		metast.ValUpdatesToVals(t, toVals(cvcs)),
		actualValidators,
	)
}

func TestCommandValidatorChangeRemoveValidator(t *testing.T) {
	const qtyVals = 2
	app, ma, cvcs := initAppCVCValidators(t, qtyVals)

	// remove a validator
	cvc := cvcs[0]
	cvc.Power = 0
	cvc.Signatures = make([]signature.Signature, 0, 1)
	cvc.Sequence = qtyVals + 1

	cvcKeys := ma[cvcKeys].([]signature.PrivateKey)
	cvc.Signatures = []signature.Signature{cvcKeys[0].Sign(cvc.SignableBytes())}

	cvcs = cvcs[1:]
	updateValidators(t, app, []CommandValidatorChange{cvc})

	actualValidators, err := app.Validators()
	require.NoError(t, err)
	metast.ValidatorsAreEquivalent(
		t,
		metast.ValUpdatesToVals(t, toVals(cvcs)),
		actualValidators,
	)
}
