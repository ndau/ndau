package ndau

import (
	metast "github.com/oneiro-ndev/metanode/pkg/meta.app/meta.state"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	"github.com/oneiro-ndev/ndaunode/pkg/ndau/backing"
	"github.com/oneiro-ndev/signature/pkg/signature"
	"github.com/pkg/errors"
)

// NewDesignateDelegate creates a new signed DesignateDelegate transaction
func NewDesignateDelegate(
	account, delegate address.Address,
	transferKey signature.PrivateKey,
) *DesignateDelegate {
	dd := &DesignateDelegate{
		Account:  account,
		Delegate: delegate,
	}
	dd.Signature = transferKey.Sign(dd.signableBytes())
	return dd
}

func (dd *DesignateDelegate) signableBytes() []byte {
	bytes := make([]byte, 0, dd.Account.Msgsize()+dd.Delegate.Msgsize())
	bytes = append(bytes, []byte(dd.Account.String())...)
	bytes = append(bytes, []byte(dd.Delegate.String())...)
	return bytes
}

// Validate implements metatx.Transactable
func (dd *DesignateDelegate) Validate(appI interface{}) error {
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
	if !acct.TransferKey.Verify(dd.signableBytes(), dd.Signature) {
		return errors.New("Invalid signature")
	}

	return nil
}

// Apply implements metatx.Transactable
func (dd *DesignateDelegate) Apply(appI interface{}) error {
	app := appI.(*App)

	app.UpdateState(func(stateI metast.State) (metast.State, error) {
		state := stateI.(*backing.State)
		as := dd.Account.String()
		ds := dd.Delegate.String()
		acct, hasAcct := state.Accounts[as]
		if !hasAcct {
			return state, errors.New("Account does not exist")
		}

		// first, remove it from its current delegate
		if acct.DelegationNode != nil {
			cs := acct.DelegationNode.String()
			currentSet, hasCurrent := state.Delegates[cs]
			if hasCurrent {
				delete(currentSet, cs)
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
	return nil
}
