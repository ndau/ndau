package tool

import (
	"fmt"
	"os"

	"github.com/oneiro-ndev/metanode/pkg/meta.app/code"
	"github.com/oneiro-ndev/metanode/pkg/meta.transaction"
	"github.com/oneiro-ndev/ndaunode/pkg/ndau"
	"github.com/pkg/errors"
	"github.com/tendermint/tendermint/rpc/client"
)

func ctkGeneric(
	node client.ABCIClient,
	ctk ndau.ChangeTransferKey,
	broadcast broadcaster,
) (interface{}, error) {
	bytes, err := metatx.TransactableToBytes(&ctk, ndau.TxIDs)
	if err != nil {
		return nil, errors.Wrap(err, "CTK failed to marshal transaction")
	}

	result, err := broadcast(node, bytes)
	if err != nil {
		if err.Error() == code.EncodingError.String() {
			fmt.Fprintf(os.Stderr, "tx bytes: %x\n", bytes)
		}
		return nil, errors.Wrap(err, "CTK failed to broadcast transaction")
	}

	return result, nil
}

// ChangeTransferKeyCommit broadcasts and commits a ChangeTransferKey transaction
func ChangeTransferKeyCommit(node client.ABCIClient, ctk ndau.ChangeTransferKey) (interface{}, error) {
	return ctkGeneric(node, ctk, broadcastCommit)
}

// ChangeTransferKeySync broadcasts a ChangeTransferKey transaction with Sync semantics
func ChangeTransferKeySync(node client.ABCIClient, ctk ndau.ChangeTransferKey) (interface{}, error) {
	return ctkGeneric(node, ctk, broadcastSync)
}

// ChangeTransferKeyAsync broadcasts a ChangeTransferKey transaction with async semantics
func ChangeTransferKeyAsync(node client.ABCIClient, ctk ndau.ChangeTransferKey) (interface{}, error) {
	return ctkGeneric(node, ctk, broadcastAsync)
}
