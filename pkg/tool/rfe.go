package tool

import (
	"github.com/oneiro-ndev/ndaunode/pkg/ndau"
	"github.com/tendermint/tendermint/rpc/client"
)

// ReleaseFromEndowmentCommit broadcasts and commits a ReleaseFromEndowment transaction
func ReleaseFromEndowmentCommit(node client.ABCIClient, rfe ndau.ReleaseFromEndowment) (interface{}, error) {
	return sendGeneric(node, &rfe, broadcastCommit, "ReleaseFromEndowment")
}

// ReleaseFromEndowmentSync broadcasts a ReleaseFromEndowment transaction with Sync semantics
func ReleaseFromEndowmentSync(node client.ABCIClient, rfe ndau.ReleaseFromEndowment) (interface{}, error) {
	return sendGeneric(node, &rfe, broadcastSync, "ReleaseFromEndowment")
}

// ReleaseFromEndowmentAsync broadcasts a ReleaseFromEndowment transaction with async semantics
func ReleaseFromEndowmentAsync(node client.ABCIClient, rfe ndau.ReleaseFromEndowment) (interface{}, error) {
	return sendGeneric(node, &rfe, broadcastAsync, "ReleaseFromEndowment")
}
