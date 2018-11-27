package ndau

import (
	"net/url"

	metast "github.com/oneiro-ndev/metanode/pkg/meta/state"
	"github.com/oneiro-ndev/ndau/pkg/ndau/backing"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	"github.com/oneiro-ndev/ndaumath/pkg/signature"
	"github.com/pkg/errors"
)

// GetAccountAddresses returns the account addresses associated with this transaction type.
func (tx *RegisterNode) GetAccountAddresses() []string {
	return []string{tx.Node.String()}
}

// NewRegisterNode creates a new RegisterNode transaction
func NewRegisterNode(
	node address.Address,
	distributionScript []byte,
	rpcAddress string,
	sequence uint64,
	keys []signature.PrivateKey,
) *RegisterNode {
	tx := &RegisterNode{
		Node:               node,
		DistributionScript: distributionScript,
		RPCAddress:         rpcAddress,
		Sequence:           sequence,
	}
	for _, key := range keys {
		tx.Signatures = append(tx.Signatures, key.Sign(tx.SignableBytes()))
	}
	return tx
}

// SignableBytes implements Transactable
func (tx *RegisterNode) SignableBytes() []byte {
	bytes := make([]byte, 0, 8+len(tx.Node.String())+len(tx.DistributionScript)+len(tx.RPCAddress))
	bytes = appendUint64(bytes, tx.Sequence)
	bytes = append(bytes, tx.Node.String()...)
	bytes = append(bytes, tx.DistributionScript...)
	bytes = append(bytes, tx.RPCAddress...)
	return bytes
}

// Validate implements metatx.Transactable
func (tx *RegisterNode) Validate(appI interface{}) error {
	if !IsChaincode(tx.DistributionScript) {
		return errors.New("DistributionScript invalid")
	}

	_, err := url.ParseRequestURI(tx.RPCAddress)
	if err != nil {
		return errors.Wrap(err, "RPCAddress invalid")
	}

	app := appI.(*App)
	state := app.GetState().(*backing.State)

	acct, _, _, err := app.getTxAccount(tx)
	if err != nil {
		return err
	}

	if acct.Stake == nil {
		return errors.New("node acct Stake is nil; node must be self-staked")
	}
	if acct.Stake.Address != tx.Node {
		return errors.New("node must be self-staked")
	}

	if state.Nodes[tx.Node.String()].Active {
		return errors.New("node is already active")
	}

	return nil
}

// Apply implements metatx.Transactable
func (tx *RegisterNode) Apply(appI interface{}) error {
	app := appI.(*App)
	err := app.applyTxDetails(tx)
	if err != nil {
		return err
	}

	return app.UpdateState(func(stateI metast.State) (metast.State, error) {
		state := stateI.(*backing.State)
		node := state.Nodes[tx.Node.String()]

		node.Active = true
		node.DistributionScript = tx.DistributionScript
		node.RPCAddress = tx.RPCAddress

		state.Nodes[tx.Node.String()] = node

		return state, nil
	})
}

// GetSource implements sourcer
func (tx *RegisterNode) GetSource(*App) (address.Address, error) {
	return tx.Node, nil
}

// GetSequence implements sequencer
func (tx *RegisterNode) GetSequence() uint64 {
	return tx.Sequence
}

// GetSignatures implements signeder
func (tx *RegisterNode) GetSignatures() []signature.Signature {
	return tx.Signatures
}

// ExtendSignatures implements Signable
func (tx *RegisterNode) ExtendSignatures(sa []signature.Signature) {
	tx.Signatures = append(tx.Signatures, sa...)
}
