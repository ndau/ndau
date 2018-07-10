package tool

import (
	"github.com/oneiro-ndev/ndaunode/pkg/ndau"
	"github.com/tendermint/tendermint/rpc/client"
)

// ComputeEAICommit broadcasts and commits a ComputeEAI transaction
func ComputeEAICommit(node client.ABCIClient, tx ndau.ComputeEAI) (interface{}, error) {
	return sendGeneric(node, &tx, broadcastCommit, "ComputeEAI")
}

// ComputeEAISync broadcasts a ComputeEAI transaction with Sync semantics
func ComputeEAISync(node client.ABCIClient, tx ndau.ComputeEAI) (interface{}, error) {
	return sendGeneric(node, &tx, broadcastSync, "ComputeEAI")
}

// ComputeEAIAsync broadcasts a ComputeEAI transaction with async semantics
func ComputeEAIAsync(node client.ABCIClient, tx ndau.ComputeEAI) (interface{}, error) {
	return sendGeneric(node, &tx, broadcastAsync, "ComputeEAI")
}
