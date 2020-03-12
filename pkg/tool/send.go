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
	metatx "github.com/ndau/metanode/pkg/meta/transaction"
	"github.com/tendermint/tendermint/rpc/client"
)

// SendCommit broadcasts and commits a transaction
func SendCommit(node client.ABCIClient, tx metatx.Transactable) (interface{}, error) {
	return sendGeneric(node, tx, broadcastCommit)
}

// SendSync broadcasts a transaction with Sync semantics
func SendSync(node client.ABCIClient, tx metatx.Transactable) (interface{}, error) {
	return sendGeneric(node, tx, broadcastSync)
}

// SendAsync broadcasts a transaction with async semantics
func SendAsync(node client.ABCIClient, tx metatx.Transactable) (interface{}, error) {
	return sendGeneric(node, tx, broadcastAsync)
}
