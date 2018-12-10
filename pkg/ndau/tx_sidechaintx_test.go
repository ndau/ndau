package ndau

import (
	"testing"

	"github.com/oneiro-ndev/chaincode/pkg/vm"
	"github.com/oneiro-ndev/metanode/pkg/meta/app/code"
	"github.com/oneiro-ndev/ndau/pkg/ndau/backing"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	"github.com/oneiro-ndev/ndaumath/pkg/signature"
	"github.com/stretchr/testify/require"
)

func TestSidechainTxSignatureMustValidate(t *testing.T) {
	app, private := initAppTx(t)
	tr := generateTransfer(t, 1, 1, []signature.PrivateKey{private})
	// I'm almost completely certain that this will be an invalid signature
	sig, err := signature.RawSignature(signature.Ed25519, make([]byte, signature.Ed25519.SignatureSize()))
	require.NoError(t, err)
	tr.Signatures = []signature.Signature{*sig}
	resp := deliverTx(t, app, tr)
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
}

func TestSidechainTxSequenceMustIncrease(t *testing.T) {
	app, private := initAppTx(t)
	invalidZero := generateTransfer(t, 1, 0, []signature.PrivateKey{private})
	resp := deliverTx(t, app, invalidZero)
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
	// valid now, because its sequence is greater than the account sequence
	tr := generateTransfer(t, 1, 1, []signature.PrivateKey{private})
	resp = deliverTx(t, app, tr)
	require.Equal(t, code.OK, code.ReturnCode(resp.Code))
	// invalid now because account sequence must have been updated
	resp = deliverTx(t, app, tr)
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
}

func TestValidationScriptValidatesSidechainTxs(t *testing.T) {
	app, private := initAppTx(t)
	public2, private2, err := signature.Generate(signature.Ed25519, nil)
	require.NoError(t, err)

	// this script just ensures that the first transfer key
	// is used, no matter how many keys are included
	script := vm.MiniAsm("handler 0 one and not enddef")

	modify(t, source, app, func(ad *backing.AccountData) {
		ad.ValidationScript = script.Bytes()
		ad.ValidationKeys = append(ad.ValidationKeys, public2)
	})

	t.Run("only first key", func(t *testing.T) {
		tr := generateTransfer(t, 123, 1, []signature.PrivateKey{private})
		resp := deliverTx(t, app, tr)
		require.Equal(t, code.OK, code.ReturnCode(resp.Code))
	})
	t.Run("both keys in order", func(t *testing.T) {
		tr := generateTransfer(t, 123, 2, []signature.PrivateKey{private, private2})
		resp := deliverTx(t, app, tr)
		require.Equal(t, code.OK, code.ReturnCode(resp.Code))
	})
	t.Run("both keys out of order", func(t *testing.T) {
		tr := generateTransfer(t, 123, 3, []signature.PrivateKey{private2, private})
		resp := deliverTx(t, app, tr)
		require.Equal(t, code.OK, code.ReturnCode(resp.Code))
	})
	t.Run("only second key", func(t *testing.T) {
		tr := generateTransfer(t, 123, 4, []signature.PrivateKey{private2})
		resp := deliverTx(t, app, tr)
		require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
	})
}

func TestSidechainTxDeductsTxFee(t *testing.T) {
	app, private := initAppTx(t)
	sA, err := address.Validate(source)
	require.NoError(t, err)

	modify(t, source, app, func(ad *backing.AccountData) {
		ad.Balance = 1
	})

	for i := 0; i < 2; i++ {
		tx, err := NewSidechainTx(
			sA,                      // address
			0,                       // sidechain id
			[]byte{0, 1, 2, 3, 4},   // tx signable bytes
			[]signature.Signature{}, // sidechain sigs
			1,                       // sequence
			[]signature.PrivateKey{private},
		)
		require.NoError(t, err)

		resp := deliverTxWithTxFee(t, app, tx)

		var expect code.ReturnCode
		if i == 0 {
			expect = code.OK
		} else {
			expect = code.InvalidTransaction
		}
		require.Equal(t, expect, code.ReturnCode(resp.Code))
	}
}
