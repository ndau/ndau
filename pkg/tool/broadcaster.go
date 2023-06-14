package tool

// ----- ---- --- -- -
// Copyright 2019 Oneiro NA, Inc. All Rights Reserved.
//
// Licensed under the Apache License 2.0 (the "License").  You may not use
// this file except in compliance with the License.  You can obtain a copy
// in the file LICENSE in the source distribution or at
// https://www.apache.org/licenses/LICENSE-2.0.txt
// - -- --- ---- -----

import (
	"fmt"
	"os"

	"github.com/ndau/metanode/pkg/meta/app/code"
	metatx "github.com/ndau/metanode/pkg/meta/transaction"
	"github.com/ndau/ndau/pkg/ndau"
	"github.com/pkg/errors"
	"github.com/oneiro-ndev/tendermint.0.32.3/rpc/client"
	ctypes "github.com/oneiro-ndev/tendermint.0.32.3/rpc/core/types"
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
	rl := ResultLog(result)
	if rl != "" {
		fmt.Fprintln(os.Stderr, rl)
	}
	if err != nil {
		if err.Error() == code.EncodingError.String() {
			fmt.Fprintf(os.Stderr, "tx bytes: %x\n", bytes)
		}
	}
	return result, errors.Wrap(err, "failed to broadcast transaction")
}

// ResultLog extracts the log message(s) from the result of a broadcast
func ResultLog(result interface{}) string {
	var out string
	if result != nil {
		switch x := result.(type) {
		case *ctypes.ResultBroadcastTxCommit:
			if x != nil {
				if x.CheckTx.Log != "" && x.DeliverTx.Log != "" {
					out = fmt.Sprintf("CheckTx: %s; DeliverTx: %s", x.CheckTx.Log, x.DeliverTx.Log)
				} else {
					out = x.CheckTx.Log + x.DeliverTx.Log
				}
			}
		case *ctypes.ResultBroadcastTx:
			if x != nil {
				out = x.Log
			}
		}
	}
	return out
}
