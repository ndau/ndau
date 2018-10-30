package tool

import (
	metatx "github.com/oneiro-ndev/metanode/pkg/meta/transaction"
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
