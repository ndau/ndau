package tool

import (
	"fmt"
	"os"

	"github.com/oneiro-ndev/metanode/pkg/meta.app/code"
	metatx "github.com/oneiro-ndev/metanode/pkg/meta.transaction"
	"github.com/oneiro-ndev/ndaunode/pkg/ndau"

	"github.com/pkg/errors"
	"github.com/tendermint/tendermint/rpc/client"
)

type broadcaster func(client.ABCIClient, []byte) (interface{}, error)

func broadcastCommit(node client.ABCIClient, tx []byte) (interface{}, error) {
	result, err := node.BroadcastTxCommit(tx)
	if err != nil {
		return result, err
	}
	rc := code.ReturnCode(result.CheckTx.Code)
	if result.CheckTx.Log != "" {
		fmt.Fprintln(os.Stderr, result.CheckTx.Log)
	}
	if result.DeliverTx.Log != "" {
		fmt.Fprintln(os.Stderr, result.DeliverTx.Log)
	}
	if rc != code.OK {
		return result, errors.New(rc.String())
	}
	return result, nil
}

func broadcastAsync(node client.ABCIClient, tx []byte) (interface{}, error) {
	result, err := node.BroadcastTxAsync(tx)
	if result.Log != "" {
		fmt.Fprintln(os.Stderr, result.Log)
	}
	if err != nil {
		return result, err
	}
	rc := code.ReturnCode(result.Code)
	if rc != code.OK {
		return result, errors.New(rc.String())
	}
	return result, nil
}

func broadcastSync(node client.ABCIClient, tx []byte) (interface{}, error) {
	result, err := node.BroadcastTxSync(tx)
	if result.Log != "" {
		fmt.Fprintln(os.Stderr, result.Log)
	}
	if err != nil {
		return result, err
	}
	rc := code.ReturnCode(result.Code)
	if rc != code.OK {
		return result, errors.New(rc.String())
	}
	return result, nil
}

func sendGeneric(
	node client.ABCIClient,
	tx metatx.Transactable,
	broadcast broadcaster,
	name string,
) (interface{}, error) {
	bytes, err := metatx.TransactableToBytes(tx, ndau.TxIDs)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("%s failed to marshal transaction", name))
	}

	result, err := broadcast(node, bytes)
	if err != nil {
		if err.Error() == code.EncodingError.String() {
			fmt.Fprintf(os.Stderr, "tx bytes: %x\n", bytes)
		}
		return nil, errors.Wrap(err, fmt.Sprintf("%s failed to broadcast transaction", name))
	}

	return result, nil
}
