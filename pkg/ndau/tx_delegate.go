package ndau

import (
	metast "github.com/oneiro-ndev/metanode/pkg/meta/state"
	"github.com/oneiro-ndev/ndau/pkg/ndau/backing"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	"github.com/oneiro-ndev/ndaumath/pkg/signature"
	"github.com/pkg/errors"
)

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
		return errors.New("Account does not exist")
	}

	return nil
}

// Apply implements metatx.Transactable
func (tx *Delegate) Apply(appI interface{}) error {
	app := appI.(*App)
	err := app.applyTxDetails(tx)
	if err != nil {
		return err
	}

	return app.UpdateState(func(stateI metast.State) (metast.State, error) {
		state := stateI.(*backing.State)
		as := tx.Target.String()
		ds := tx.Node.String()
		acct, hasAcct := state.GetAccount(tx.Target, app.blockTime)
		if !hasAcct {
			return state, errors.New("Account does not exist")
		}

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
		acct.DelegationNode = &tx.Node
		state.Accounts[as] = acct

		// update the target delegate's set
		dest := state.Delegates[ds]
		if dest == nil {
			dest = make(map[string]struct{})
		}
		dest[as] = struct{}{}
		state.Delegates[ds] = dest

		return state, nil
	})
}

// GetSource implements sourcer
func (tx *Delegate) GetSource(*App) (address.Address, error) {
	return tx.Target, nil
}

// GetSequence implements sequencer
func (tx *Delegate) GetSequence() uint64 {
	return tx.Sequence
}

// GetSignatures implements signeder
func (tx *Delegate) GetSignatures() []signature.Signature {
	return tx.Signatures
}

// ExtendSignatures implements Signable
func (tx *Delegate) ExtendSignatures(sa []signature.Signature) {
	tx.Signatures = append(tx.Signatures, sa...)
}
