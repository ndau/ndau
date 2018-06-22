package tool

import (
	"github.com/oneiro-ndev/ndaunode/pkg/ndau"
	"github.com/tendermint/tendermint/rpc/client"
)

// ChangeTransferKeyCommit broadcasts and commits a ChangeTransferKey transaction
func ChangeTransferKeyCommit(node client.ABCIClient, ctk ndau.ChangeTransferKey) (interface{}, error) {
	return sendGeneric(node, &ctk, broadcastCommit, "ChangeTransferKey")
}

// ChangeTransferKeySync broadcasts a ChangeTransferKey transaction with Sync semantics
func ChangeTransferKeySync(node client.ABCIClient, ctk ndau.ChangeTransferKey) (interface{}, error) {
	return sendGeneric(node, &ctk, broadcastSync, "ChangeTransferKey")
}

// ChangeTransferKeyAsync broadcasts a ChangeTransferKey transaction with async semantics
func ChangeTransferKeyAsync(node client.ABCIClient, ctk ndau.ChangeTransferKey) (interface{}, error) {
	return sendGeneric(node, &ctk, broadcastAsync, "ChangeTransferKey")
}
