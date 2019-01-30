package ndau

import (
	metast "github.com/oneiro-ndev/metanode/pkg/meta/state"
	"github.com/oneiro-ndev/ndau/pkg/ndau/backing"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	"github.com/oneiro-ndev/ndaumath/pkg/signature"
	"github.com/pkg/errors"
)

// Validate implements metatx.Transactable
func (tx *UnregisterNode) Validate(appI interface{}) error {

	app := appI.(*App)
	state := app.GetState().(*backing.State)

	_, _, _, err := app.getTxAccount(tx)
	if err != nil {
		return err
	}

	_, isNode := state.Nodes[tx.Node.String()]
	if !isNode {
		return errors.New("not a node")
	}

	return nil
}

// Apply implements metatx.Transactable
func (tx *UnregisterNode) Apply(appI interface{}) error {
	app := appI.(*App)
	err := app.applyTxDetails(tx)
	if err != nil {
		return err
	}

	return app.UpdateState(func(stateI metast.State) (metast.State, error) {
		state := stateI.(*backing.State)

		node, ok := state.Nodes[tx.Node.String()]
		if !ok {
			// empty node, no costakers
			return state, nil
		}

		for costaker := range node.Costakers {
			csA, err := address.Validate(costaker)
			if err != nil {
				// whatever, don't hang around dealing with this
				continue
			}
			cs, exists := state.GetAccount(csA, app.blockTime)
			if !exists {
				// oh well, it must have been removed for no balance or something sometime
				continue
			}
			cs.Stake = nil
			state.Accounts[costaker] = cs
		}

		// don't just deactivate but delete this node
		delete(state.Nodes, tx.Node.String())

		nodeAD, exists := state.GetAccount(tx.Node, app.blockTime)
		if exists {
			nodeAD.Stake = nil
			state.Accounts[tx.Node.String()] = nodeAD
		}

		return state, nil
	})
}

// GetSource implements sourcer
func (tx *UnregisterNode) GetSource(*App) (address.Address, error) {
	return tx.Node, nil
}

// GetSequence implements sequencer
func (tx *UnregisterNode) GetSequence() uint64 {
	return tx.Sequence
}

// GetSignatures implements signeder
func (tx *UnregisterNode) GetSignatures() []signature.Signature {
	return tx.Signatures
}

// ExtendSignatures implements Signable
func (tx *UnregisterNode) ExtendSignatures(sa []signature.Signature) {
	tx.Signatures = append(tx.Signatures, sa...)
}

// GetAccountAddresses returns the account addresses associated with this transaction type.
func (tx *UnregisterNode) GetAccountAddresses() []string {
	return []string{tx.Node.String()}
}
