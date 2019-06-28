package ndau

import (
	metast "github.com/oneiro-ndev/metanode/pkg/meta/state"
	"github.com/oneiro-ndev/ndau/pkg/ndau/backing"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	"github.com/oneiro-ndev/ndaumath/pkg/signature"
	sv "github.com/oneiro-ndev/system_vars/pkg/system_vars"
	"github.com/pkg/errors"
)

// GetAccountAddresses returns the account addresses associated with this transaction type.
func (tx *ReleaseFromEndowment) GetAccountAddresses() []string {
	return []string{tx.Destination.String()}
}

// Validate implements metatx.Transactable
func (tx *ReleaseFromEndowment) Validate(appI interface{}) error {
	app := appI.(*App)

	if tx.Qty <= 0 {
		return errors.New("RFE qty may not be <= 0")
	}

	_, _, _, err := app.getTxAccount(tx)

	return err
}

// Apply implements metatx.Transactable
func (tx *ReleaseFromEndowment) Apply(appI interface{}) error {
	app := appI.(*App)
	return app.UpdateState(app.applyTxDetails(tx), func(stateI metast.State) (metast.State, error) {
		var err error
		state := stateI.(*backing.State)

		acct, _ := app.getAccount(tx.Destination)
		acct.Balance, err = acct.Balance.Add(tx.Qty)
		if err != nil {
			return state, err
		}
		acct.UpdateCurrencySeat(app.BlockTime())
		state.Accounts[tx.Destination.String()] = acct

		// we give up overflow protection here in exchange for error-free
		// operation; we have external constraints that we will never issue
		// more than (30 million) * (100 million) napu, or 0.03% of 64-bits,
		// so this should be fine
		state.TotalRFE += tx.Qty

		return state, err
	})
}

// GetSource implements Sourcer
func (tx *ReleaseFromEndowment) GetSource(app *App) (addr address.Address, err error) {
	err = app.System(sv.ReleaseFromEndowmentAddressName, &addr)
	return
}

// GetSequence implements Sequencer
func (tx *ReleaseFromEndowment) GetSequence() uint64 {
	return tx.Sequence
}

// GetSignatures implements Signeder
func (tx *ReleaseFromEndowment) GetSignatures() []signature.Signature {
	return tx.Signatures
}

// ExtendSignatures implements Signable
func (tx *ReleaseFromEndowment) ExtendSignatures(sa []signature.Signature) {
	tx.Signatures = append(tx.Signatures, sa...)
}
