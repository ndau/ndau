package ndau

import (
	"encoding/binary"

	metast "github.com/oneiro-ndev/metanode/pkg/meta/state"
	"github.com/oneiro-ndev/ndau/pkg/ndau/backing"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	"github.com/oneiro-ndev/signature/pkg/signature"
	"github.com/pkg/errors"
)

// NewDelegate creates a new signed Delegate transaction
func NewDelegate(
	account, delegate address.Address,
	sequence uint64,
	keys []signature.PrivateKey,
) *Delegate {
	dd := &Delegate{
		Target:   account,
		Node:     delegate,
		Sequence: sequence,
	}
	for _, key := range keys {
		dd.Signatures = append(dd.Signatures, key.Sign(dd.SignableBytes()))
	}
	return dd
}

// SignableBytes implements Transactable
func (dd *Delegate) SignableBytes() []byte {
	bytes := make([]byte, 8, dd.Target.Msgsize()+dd.Node.Msgsize()+8)
	binary.BigEndian.PutUint64(bytes, dd.Sequence)
	bytes = append(bytes, []byte(dd.Target.String())...)
	bytes = append(bytes, []byte(dd.Node.String())...)
	return bytes
}

// Validate implements metatx.Transactable
func (dd *Delegate) Validate(appI interface{}) error {
	_, err := address.Validate(dd.Target.String())
	if err != nil {
		return errors.Wrap(err, "Account")
	}
	_, err = address.Validate(dd.Node.String())
	if err != nil {
		return errors.Wrap(err, "Delegate")
	}

	app := appI.(*App)
	_, hasAccount, err := app.GetState().(*backing.State).GetValidAccount(
		dd.Target,
		app.blockTime,
		dd.Sequence,
		dd.SignableBytes(),
		dd.Signatures,
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
func (dd *Delegate) Apply(appI interface{}) error {
	app := appI.(*App)

	return app.UpdateState(func(stateI metast.State) (metast.State, error) {
		state := stateI.(*backing.State)
		as := dd.Target.String()
		ds := dd.Node.String()
		acct, hasAcct := state.GetAccount(dd.Target, app.blockTime)
		if !hasAcct {
			return state, errors.New("Account does not exist")
		}

		acct.Sequence = dd.Sequence

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
		acct.DelegationNode = &dd.Node
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
