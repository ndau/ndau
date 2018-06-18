package ndau

import (
	"encoding/hex"
	"testing"

	"github.com/oneiro-ndev/metanode/pkg/meta.app/code"
	tx "github.com/oneiro-ndev/metanode/pkg/meta.transaction"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	"github.com/oneiro-ndev/signature/pkg/signature"
	"github.com/stretchr/testify/require"

	"github.com/oneiro-ndev/ndaunode/pkg/ndau/backing"
	abci "github.com/tendermint/abci/types"
)

// define ownership key and address for target
const (
	targetPrivateHex = "995db0d8fcdbac19f1d7ebb3c320ac41749b650878ed6fbe60f59364bf24e96ecc9db61c8cfaaf4bac2e64fc6ba064e53db7d5d9d121593d5d7e1e4ab3a4b835"
	targetPublicHex  = "cc9db61c8cfaaf4bac2e64fc6ba064e53db7d5d9d121593d5d7e1e4ab3a4b835"
	target           = "ndapm8etnfe53nc3sbpcnkpwnc4hku7ahbfr2um7mwupttgt"
)

var (
	targetPrivate *signature.PrivateKey
	targetPublic  *signature.PublicKey
)

func init() {
	tPrivBytes, err := hex.DecodeString(targetPrivateHex)
	if err != nil {
		panic(err)
	}
	targetPrivate, err = signature.RawPrivateKey(signature.Ed25519, tPrivBytes)
	if err != nil {
		panic(err)
	}
	tPubBytes, err := hex.DecodeString(targetPublicHex)
	if err != nil {
		panic(err)
	}
	targetPublic, err = signature.RawPublicKey(signature.Ed25519, tPubBytes)
	if err != nil {
		panic(err)
	}
}

func initAppCTK(t *testing.T) *App {
	app := initApp(t)
	app.InitChain(abci.RequestInitChain{})

	// this ensures the target address exists
	modify(t, target, app, func(acct *backing.AccountData) {})

	return app
}

func TestCTKAddressFieldValidates(t *testing.T) {
	// flip the bits of the last byte so the address is no longer correct
	addrBytes := []byte(target)
	addrBytes[len(addrBytes)-1] = ^addrBytes[len(addrBytes)-1]
	addrS := string(addrBytes)

	// ensure that we didn't accidentally create a valid address
	addr, err := address.Validate(addrS)
	require.Error(t, err)

	newPublic, _, err := signature.Generate(signature.Ed25519, nil)

	// the address is invalid, but newCTK doesn't validate this
	ctk := NewChangeTransferKey(addr, newPublic, *targetPublic, SigningKeyOwnership, *targetPrivate)

	// However, the resultant transaction must not be valid
	ctkBytes, err := tx.TransactableToBytes(&ctk, TxIDs)
	require.NoError(t, err)

	app := initAppCTK(t)
	resp := app.CheckTx(ctkBytes)
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
}
