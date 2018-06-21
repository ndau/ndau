package tool

import (
	"errors"

	"github.com/oneiro-ndev/metanode/pkg/meta.app/code"
	"github.com/tendermint/tendermint/rpc/client"
)

type broadcaster func(client.ABCIClient, []byte) (interface{}, error)

func broadcastCommit(node client.ABCIClient, tx []byte) (interface{}, error) {
	result, err := node.BroadcastTxCommit(tx)
	if err != nil {
		return result, err
	}
	rc := code.ReturnCode(result.CheckTx.Code)
	if rc != code.OK {
		return result, errors.New(rc.String())
	}
	return result, nil
}

func broadcastAsync(node client.ABCIClient, tx []byte) (interface{}, error) {
	result, err := node.BroadcastTxAsync(tx)
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
	if err != nil {
		return result, err
	}
	rc := code.ReturnCode(result.Code)
	if rc != code.OK {
		return result, errors.New(rc.String())
	}
	return result, nil
}
