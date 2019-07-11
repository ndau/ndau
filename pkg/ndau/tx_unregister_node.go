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

	return app.UpdateState(app.applyTxDetails(tx), func(stateI metast.State) (metast.State, error) {
		state := stateI.(*backing.State)
		delete(state.Nodes, tx.Node.String())
		return state, nil
	})
}

// GetSource implements Sourcer
func (tx *UnregisterNode) GetSource(*App) (address.Address, error) {
	return tx.Node, nil
}

// GetSequence implements Sequencer
func (tx *UnregisterNode) GetSequence() uint64 {
	return tx.Sequence
}

// GetSignatures implements Signeder
func (tx *UnregisterNode) GetSignatures() []signature.Signature {
	return tx.Signatures
}

// ExtendSignatures implements Signable
func (tx *UnregisterNode) ExtendSignatures(sa []signature.Signature) {
	tx.Signatures = append(tx.Signatures, sa...)
}
