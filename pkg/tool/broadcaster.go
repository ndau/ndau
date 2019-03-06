package tool

import (
	"fmt"
	"os"
	"reflect"

	"github.com/oneiro-ndev/metanode/pkg/meta/app/code"
	metatx "github.com/oneiro-ndev/metanode/pkg/meta/transaction"
	"github.com/oneiro-ndev/ndau/pkg/ndau"
	"github.com/pkg/errors"
	"github.com/tendermint/tendermint/rpc/client"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
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
) (interface{}, error) {
	bytes, err := metatx.Marshal(tx, ndau.TxIDs)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal transaction")
	}

	result, err := broadcast(node, bytes)
	if err != nil {
		if err.Error() == code.EncodingError.String() {
			fmt.Fprintf(os.Stderr, "tx bytes: %x\n", bytes)
		}
		return nil, errors.Wrap(err, "failed to broadcast transaction")
	}

	return result, nil
}

// ResultLog extracts the log message(s) from the result of a broadcast
func ResultLog(result interface{}) string {
	var out string
	switch x := result.(type) {
	case *ctypes.ResultBroadcastTxCommit:
		if x.CheckTx.Log != "" && x.DeliverTx.Log != "" {
			out = fmt.Sprintf("CheckTx: %s; DeliverTx: %s", x.CheckTx.Log, x.DeliverTx.Log)
		} else {
			out = x.CheckTx.Log + x.DeliverTx.Log
		}
	case *ctypes.ResultBroadcastTx:
		out = x.Log
	default:
		out = fmt.Sprintf("can't extract log from %s", reflect.TypeOf(result))
	}
	return out
}
