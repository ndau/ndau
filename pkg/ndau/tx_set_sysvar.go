package ndau

import (
	"fmt"

	metast "github.com/oneiro-ndev/metanode/pkg/meta/state"
	"github.com/oneiro-ndev/ndau/pkg/ndau/backing"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	"github.com/oneiro-ndev/ndaumath/pkg/signature"
	sv "github.com/oneiro-ndev/system_vars/pkg/system_vars"
	"github.com/pkg/errors"
)

// Validate implements metatx.Transactable
func (tx *SetSysvar) Validate(appI interface{}) error {
	app := appI.(*App)

	// if we let someone overwrite the sysvar governing who is allowed to
	// set the sysvar with bad data, then we're hosed. Let's ensure that
	// if that's the sysvar being set, it's by an account which has been
	// claimed.
	if tx.Name == sv.SetSysvarAddressName {
		var acct address.Address
		leftovers, err := acct.UnmarshalMsg(tx.Value)
		if err != nil {
			return errors.Wrap(err,
				fmt.Sprintf(
					"value for %s must be a valid Address",
					sv.SetSysvarAddressName,
				),
			)
		}
		if len(leftovers) > 0 {
			return fmt.Errorf(
				"value for %s must not have leftovers; got %x",
				sv.SetSysvarAddressName,
				leftovers,
			)
		}

		data, exists := app.getAccount(acct)
		if !exists {
			return fmt.Errorf(
				"new %s must be an account which exists; %s doesn't",
				sv.SetSysvarAddressName,
				acct,
			)
		}

		if len(data.ValidationKeys) == 0 {
			return fmt.Errorf(
				"new %s acct (%s) must have at least 1 validation key",
				sv.SetSysvarAddressName,
				acct,
			)
		}
	}

	_, _, _, err := app.getTxAccount(tx)

	return err
}

// Apply implements metatx.Transactable
func (tx *SetSysvar) Apply(appI interface{}) error {
	app := appI.(*App)
	err := app.applyTxDetails(tx)
	if err != nil {
		return err
	}

	return app.UpdateState(func(stateI metast.State) (metast.State, error) {
		state := stateI.(*backing.State)
		state.Sysvars[tx.Name] = tx.Value
		return state, err
	})
}

// GetSource implements Sourcer
func (tx *SetSysvar) GetSource(app *App) (addr address.Address, err error) {
	err = app.System(sv.SetSysvarAddressName, &addr)
	if err != nil {
		return
	}
	if addr.Revalidate() != nil {
		err = fmt.Errorf(
			"%s sysvar not properly set; SetSysvar therefore disallowed",
			sv.SetSysvarAddressName,
		)
		return
	}
	return
}

// GetSequence implements Sequencer
func (tx *SetSysvar) GetSequence() uint64 {
	return tx.Sequence
}

// GetSignatures implements Signeder
func (tx *SetSysvar) GetSignatures() []signature.Signature {
	return tx.Signatures
}

// ExtendSignatures implements Signable
func (tx *SetSysvar) ExtendSignatures(sa []signature.Signature) {
	tx.Signatures = append(tx.Signatures, sa...)
}

// GetAccountAddresses returns the account addresses associated with this transaction type.
func (tx *SetSysvar) GetAccountAddresses() []string {
	return []string{}
}

// GetName implements SysvarIndexable.
func (tx *SetSysvar) GetName() string {
	return tx.Name
}

// GetValue implements SysvarIndexable.
func (tx *SetSysvar) GetValue() []byte {
	return tx.Value
}
