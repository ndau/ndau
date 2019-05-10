package ndau

import (
	"fmt"

	metast "github.com/oneiro-ndev/metanode/pkg/meta/state"
	"github.com/oneiro-ndev/ndau/pkg/ndau/backing"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	"github.com/oneiro-ndev/ndaumath/pkg/pricecurve"
	"github.com/oneiro-ndev/ndaumath/pkg/signature"
	sv "github.com/oneiro-ndev/system_vars/pkg/system_vars"
	"github.com/pkg/errors"
)

// return a function intended to be run within app.UpdateState
//
// special case: if the input is negative, just use the existing value
// this is used to allow updating market price without affecting NAV
func (app *App) updateNAVAndSIB(nav pricecurve.Nanocent) func(stateI metast.State) (metast.State, error) {
	return func(stateI metast.State) (metast.State, error) {
		state := stateI.(*backing.State)
		state.EndowmentNAV = nav
		sib, target, err := app.calculateCurrentSIB(-1, nav)
		if err != nil {
			return stateI, err
		}
		state.SIB = sib
		state.TargetPrice = target

		return state, err
	}
}

// Validate implements metatx.Transactable
func (tx *RecordEndowmentNAV) Validate(appI interface{}) error {
	app := appI.(*App)

	if tx.NAV <= 0 {
		return errors.New("RecordEndowmentNAV NAV may not be <= 0")
	}

	_, _, _, err := app.getTxAccount(tx)

	return err
}

// Apply implements metatx.Transactable
func (tx *RecordEndowmentNAV) Apply(appI interface{}) error {
	app := appI.(*App)
	err := app.applyTxDetails(tx)
	if err != nil {
		return err
	}

	return app.UpdateState(app.updateNAVAndSIB(tx.NAV))
}

// GetSource implements sourcer
func (tx *RecordEndowmentNAV) GetSource(app *App) (addr address.Address, err error) {
	err = app.System(sv.RecordEndowmentNAVAddressName, &addr)
	if err != nil {
		return
	}
	if addr.Revalidate() != nil {
		err = fmt.Errorf(
			"%s sysvar not set; RecordEndowmentNAV therefore disallowed",
			sv.RecordEndowmentNAVAddressName,
		)
		return
	}
	return
}

// GetSequence implements sequencer
func (tx *RecordEndowmentNAV) GetSequence() uint64 {
	return tx.Sequence
}

// GetSignatures implements signeder
func (tx *RecordEndowmentNAV) GetSignatures() []signature.Signature {
	return tx.Signatures
}

// ExtendSignatures implements Signable
func (tx *RecordEndowmentNAV) ExtendSignatures(sa []signature.Signature) {
	tx.Signatures = append(tx.Signatures, sa...)
}

// GetAccountAddresses returns the account addresses associated with this transaction type.
func (tx *RecordEndowmentNAV) GetAccountAddresses() []string {
	return []string{}
}
