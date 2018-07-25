package tool

import (
	"fmt"

	"github.com/oneiro-ndev/metanode/pkg/meta/app/code"
	metatx "github.com/oneiro-ndev/metanode/pkg/meta/transaction"
	"github.com/oneiro-ndev/ndau/pkg/ndau"
	"github.com/pkg/errors"
	"github.com/tendermint/tendermint/crypto/ed25519"
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
	if len(key) != ed25519.PubKeyEd25519Size {
		return nil, errors.New("Invalid size for ed25519 public key")
	}
	pk := ed25519.PubKeyEd25519{}
	copy(pk[:], key)

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
