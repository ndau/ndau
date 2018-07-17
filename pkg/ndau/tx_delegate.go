package ndau

import (
	"encoding/binary"

	metast "github.com/oneiro-ndev/metanode/pkg/meta.app/meta.state"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	"github.com/oneiro-ndev/ndau/pkg/ndau/backing"
	"github.com/oneiro-ndev/signature/pkg/signature"
	"github.com/pkg/errors"
)

// NewDelegate creates a new signed Delegate transaction
func NewDelegate(
	account, delegate address.Address,
	sequence uint64,
	transferKey signature.PrivateKey,
) *Delegate {
	dd := &Delegate{
		Account:  account,
		Delegate: delegate,
		Sequence: sequence,
	}
	dd.Signature = transferKey.Sign(dd.signableBytes())
	return dd
}

func (dd *Delegate) signableBytes() []byte {
	bytes := make([]byte, 8, dd.Account.Msgsize()+dd.Delegate.Msgsize()+8)
	binary.BigEndian.PutUint64(bytes, dd.Sequence)
	bytes = append(bytes, []byte(dd.Account.String())...)
	bytes = append(bytes, []byte(dd.Delegate.String())...)
	return bytes
}

// Validate implements metatx.Transactable
func (dd *Delegate) Validate(appI interface{}) error {
	_, err := address.Validate(dd.Account.String())
	if err != nil {
		return errors.Wrap(err, "Account")
	}
	_, err = address.Validate(dd.Delegate.String())
	if err != nil {
		return errors.Wrap(err, "Delegate")
	}

	app := appI.(*App)
	acct, hasAcct := app.GetState().(*backing.State).Accounts[dd.Account.String()]
	if !hasAcct {
		return errors.New("Account does not exist")
	}
	if dd.Sequence <= acct.Sequence {
		return errors.New("Sequence too low")
	}
	if acct.TransferKey == nil {
		return errors.New("Transfer key not set")
	}
	if !acct.TransferKey.Verify(dd.signableBytes(), dd.Signature) {
		return errors.New("Invalid signature")
	}

	return nil
}

// Apply implements metatx.Transactable
func (dd *Delegate) Apply(appI interface{}) error {
	app := appI.(*App)

	return app.UpdateState(func(stateI metast.State) (metast.State, error) {
		state := stateI.(*backing.State)
		as := dd.Account.String()
		ds := dd.Delegate.String()
		acct, hasAcct := state.Accounts[as]
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
		acct.DelegationNode = &dd.Delegate
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
