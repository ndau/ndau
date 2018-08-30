package ndau

import (
	metast "github.com/oneiro-ndev/metanode/pkg/meta/state"
	"github.com/oneiro-ndev/ndau/pkg/ndau/backing"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	math "github.com/oneiro-ndev/ndaumath/pkg/types"
	"github.com/oneiro-ndev/signature/pkg/signature"
	"github.com/pkg/errors"
)

// NewDelegate creates a new signed Delegate transaction
func NewDelegate(
	account, delegate address.Address,
	sequence uint64,
	keys []signature.PrivateKey,
) *Delegate {
	tx := &Delegate{
		Target:   account,
		Node:     delegate,
		Sequence: sequence,
	}
	for _, key := range keys {
		tx.Signatures = append(tx.Signatures, key.Sign(tx.SignableBytes()))
	}
	return tx
}

// SignableBytes implements Transactable
func (tx *Delegate) SignableBytes() []byte {
	bytes := make([]byte, 0, tx.Target.Msgsize()+tx.Node.Msgsize()+8)
	bytes = appendUint64(bytes, tx.Sequence)
	bytes = append(bytes, []byte(tx.Target.String())...)
	bytes = append(bytes, []byte(tx.Node.String())...)
	return bytes
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
	_, hasAccount, _, err := app.getTxAccount(
		tx,
		tx.Target,
		tx.Sequence,
		tx.Signatures,
	)
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

	return app.UpdateState(func(stateI metast.State) (metast.State, error) {
		state := stateI.(*backing.State)
		as := tx.Target.String()
		ds := tx.Node.String()
		acct, hasAcct := state.GetAccount(tx.Target, app.blockTime)
		if !hasAcct {
			return state, errors.New("Account does not exist")
		}

		acct.Sequence = tx.Sequence

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

// CalculateTxFee implements metatx.Transactable
func (tx *Delegate) CalculateTxFee(appI interface{}) (math.Ndau, error) {
	app := appI.(*App)
	return app.calculateTxFee(tx)
}
