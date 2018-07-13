package tool

import (
	"github.com/oneiro-ndev/ndaunode/pkg/ndau"
	"github.com/tendermint/tendermint/rpc/client"
)

// DelegateCommit broadcasts and commits a Delegate transaction
func DelegateCommit(node client.ABCIClient, tx ndau.Delegate) (interface{}, error) {
	return sendGeneric(node, &tx, broadcastCommit, "Delegate")
}

// DelegateSync broadcasts a Delegate transaction with Sync semantics
func DelegateSync(node client.ABCIClient, tx ndau.Delegate) (interface{}, error) {
	return sendGeneric(node, &tx, broadcastSync, "Delegate")
}

// DelegateAsync broadcasts a Delegate transaction with async semantics
func DelegateAsync(node client.ABCIClient, tx ndau.Delegate) (interface{}, error) {
	return sendGeneric(node, &tx, broadcastAsync, "Delegate")
}
