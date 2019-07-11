package ndau

import (
	metast "github.com/oneiro-ndev/metanode/pkg/meta/state"
	"github.com/oneiro-ndev/ndau/pkg/ndau/backing"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	"github.com/oneiro-ndev/ndaumath/pkg/signature"
	sv "github.com/oneiro-ndev/system_vars/pkg/system_vars"
	"github.com/pkg/errors"
)

// Validate implements metatx.Transactable
func (tx *Issue) Validate(appI interface{}) error {
	app := appI.(*App)

	if tx.Qty <= 0 {
		return errors.New("Issue qty may not be <= 0")
	}

	state := app.GetState().(*backing.State)
	if state.TotalIssue+tx.Qty > state.TotalRFE {
		return errors.New("cannot issue more ndau than have been RFE'd")
	}

	_, _, _, err := app.getTxAccount(tx)

	return err
}

// Apply implements metatx.Transactable
func (tx *Issue) Apply(appI interface{}) error {
	app := appI.(*App)

	return app.UpdateState(
		app.applyTxDetails(tx),
		func(stateI metast.State) (metast.State, error) {
			state := stateI.(*backing.State)

			// we give up overflow protection here in exchange for error-free
			// operation; we have external constraints that we will never issue
			// more than (30 million) * (100 million) napu, or 0.03% of 64-bits,
			// so this should be fine
			state.TotalIssue += tx.Qty
			return state, nil
		},
		app.updatePricesAndSIB(-1),
	)
}

// GetSource implements Sourcer
func (tx *Issue) GetSource(app *App) (addr address.Address, err error) {
	err = app.System(sv.ReleaseFromEndowmentAddressName, &addr)
	return
}

// GetSequence implements Sequencer
func (tx *Issue) GetSequence() uint64 {
	return tx.Sequence
}

// GetSignatures implements Signeder
func (tx *Issue) GetSignatures() []signature.Signature {
	return tx.Signatures
}

// ExtendSignatures implements Signable
func (tx *Issue) ExtendSignatures(sa []signature.Signature) {
	tx.Signatures = append(tx.Signatures, sa...)
}
