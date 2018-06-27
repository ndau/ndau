package tool

import (
	"github.com/oneiro-ndev/ndaunode/pkg/ndau"
	"github.com/tendermint/tendermint/rpc/client"
)

// TransferCommit broadcasts and commits a Transfer transaction
func TransferCommit(node client.ABCIClient, tr ndau.Transfer) (interface{}, error) {
	return sendGeneric(node, &tr, broadcastCommit, "Transfer")
}

// TransferSync broadcasts a Transfer transaction with Sync semantics
func TransferSync(node client.ABCIClient, tr ndau.Transfer) (interface{}, error) {
	return sendGeneric(node, &tr, broadcastSync, "Transfer")
}

// TransferAsync broadcasts a Transfer transaction with async semantics
func TransferAsync(node client.ABCIClient, tr ndau.Transfer) (interface{}, error) {
	return sendGeneric(node, &tr, broadcastAsync, "Transfer")
}
