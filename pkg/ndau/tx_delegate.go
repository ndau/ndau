package ndau

import (
	metast "github.com/oneiro-ndev/metanode/pkg/meta/state"
	"github.com/oneiro-ndev/ndau/pkg/ndau/backing"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	"github.com/oneiro-ndev/ndaumath/pkg/signature"
	"github.com/pkg/errors"
)

// Delegate Target to Node
func (app *App) Delegate(state *backing.State, target, node address.Address) error {
	as := target.String()
	ds := node.String()
	acct, _ := app.getAccount(target)

	// remove it from its current delegate
	if acct.DelegationNode != nil {
		cs := acct.DelegationNode.String()
		currentSet, hasCurrent := state.Delegates[cs]
		if hasCurrent {
			delete(currentSet, as)
			state.Delegates[cs] = currentSet
		}
	}

	// set its delegate
	acct.DelegationNode = &node
	state.Accounts[as] = acct

	// update the target delegate's set
	dest := state.Delegates[ds]
	if dest == nil {
		dest = make(map[string]struct{})
	}
	dest[as] = struct{}{}
	state.Delegates[ds] = dest

	return nil
}

// GetAccountAddresses returns the account addresses associated with this transaction type.
func (tx *Delegate) GetAccountAddresses() []string {
	return []string{tx.Target.String(), tx.Node.String()}
}

// Validate implements metatx.Transactable
func (tx *Delegate) Validate(appI interface{}) error {
	_, err := address.Validate(tx.Target.String())
	if err != nil {
		return errors.Wrap(err, "Account")
	}
	_, err = address.Validate(tx.Node.String())
	if err != nil {
		return errors.Wrap(err, "Delegate")
	}

	app := appI.(*App)
	_, hasAccount, _, err := app.getTxAccount(tx)
	if err != nil {
		return err
	}
	if !hasAccount {
		return errors.New("delegation target account does not exist")
	}
	// delegated node must be active
	if app.IsFeatureActive("NodeActiveCheck") && !app.GetState().(*backing.State).IsActiveNode(tx.Node) {
		return errors.New("node must be active and is not")
	}

	return nil
}

// Apply implements metatx.Transactable
func (tx *Delegate) Apply(appI interface{}) error {
	app := appI.(*App)

	return app.UpdateState(app.applyTxDetails(tx), func(stI metast.State) (metast.State, error) {
		state := stI.(*backing.State)
		err := app.Delegate(state, tx.Target, tx.Node)
		return state, err
	})
}

// GetSource implements Sourcer
func (tx *Delegate) GetSource(*App) (address.Address, error) {
	return tx.Target, nil
}

// GetSequence implements Sequencer
func (tx *Delegate) GetSequence() uint64 {
	return tx.Sequence
}

// GetSignatures implements Signeder
func (tx *Delegate) GetSignatures() []signature.Signature {
	return tx.Signatures
}

// ExtendSignatures implements Signable
func (tx *Delegate) ExtendSignatures(sa []signature.Signature) {
	tx.Signatures = append(tx.Signatures, sa...)
}
