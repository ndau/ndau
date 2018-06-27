package tool

import (
	"github.com/oneiro-ndev/ndaunode/pkg/ndau"
	"github.com/tendermint/tendermint/rpc/client"
)

// ChangeEscrowPeriodCommit broadcasts and commits a ChangeEscrowPeriod transaction
func ChangeEscrowPeriodCommit(node client.ABCIClient, tr ndau.ChangeEscrowPeriod) (interface{}, error) {
	return sendGeneric(node, &tr, broadcastCommit, "ChangeEscrowPeriod")
}

// ChangeEscrowPeriodSync broadcasts a ChangeEscrowPeriod transaction with Sync semantics
func ChangeEscrowPeriodSync(node client.ABCIClient, tr ndau.ChangeEscrowPeriod) (interface{}, error) {
	return sendGeneric(node, &tr, broadcastSync, "ChangeEscrowPeriod")
}

// ChangeEscrowPeriodAsync broadcasts a ChangeEscrowPeriod transaction with async semantics
func ChangeEscrowPeriodAsync(node client.ABCIClient, tr ndau.ChangeEscrowPeriod) (interface{}, error) {
	return sendGeneric(node, &tr, broadcastAsync, "ChangeEscrowPeriod")
}
