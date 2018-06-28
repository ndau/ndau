package tool

import (
	"fmt"

	"github.com/oneiro-ndev/metanode/pkg/meta.app/code"
	metatx "github.com/oneiro-ndev/metanode/pkg/meta.transaction"
	"github.com/oneiro-ndev/ndaunode/pkg/ndau"
	crypto "github.com/tendermint/go-crypto"
	"github.com/tendermint/tendermint/rpc/client"
	rpctypes "github.com/tendermint/tendermint/rpc/core/types"
)

// GTVC sends a globally trusted validator change to the specified node
//
// Globally trusted validator changes are a debugging tool, and will be
// banned at some point pre-1.0.
func GTVC(node client.ABCIClient, key []byte, power int64) (
	*rpctypes.ResultBroadcastTx, error,
) {
	// we expect ed25519 public key inputs, and tendermint expects a
	// type byte prefix on keys, so let's append that

	ekb := append(
		[]byte{0x01}, // see https://github.com/tendermint/go-crypto/blob/915416979bf70efa4bcbf1c6cd5d64c5fff9fc19/keys/keys.go#L14-L15
		key...,
	)

	// validate the given public key
	pk, err := crypto.PubKeyFromBytes(ekb)
	if err != nil {
		return nil, err
	}

	gtvcTxID := metatx.TxIDMap{
		metatx.TxID(0xff): &ndau.GTValidatorChange{},
	}

	gtvcb, err := metatx.Marshal(&ndau.GTValidatorChange{
		PublicKey: pk.Bytes(),
		Power:     power,
	}, gtvcTxID)

	if err != nil {
		return nil, err
	}

	result, err := node.BroadcastTxAsync(gtvcb)
	if err != nil {
		return nil, err
	}

	rc := code.ReturnCode(result.Code)
	if rc != code.OK {
		return result, fmt.Errorf(
			"GTVC returned code: %s (%s)",
			rc.String(), result.Log,
		)
	}

	return result, nil
}
