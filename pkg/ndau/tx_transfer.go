package ndau

import (
	metast "github.com/oneiro-ndev/metanode/pkg/meta/state"
	"github.com/oneiro-ndev/ndau/pkg/ndau/backing"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	"github.com/oneiro-ndev/ndaumath/pkg/signature"
	math "github.com/oneiro-ndev/ndaumath/pkg/types"
	"github.com/pkg/errors"
)

// GetAccountAddresses returns the account addresses associated with this transaction type.
func (tx *Transfer) GetAccountAddresses() []string {
	return []string{tx.Source.String(), tx.Destination.String()}
}

// Validate satisfies metatx.Transactable
func (tx *Transfer) Validate(appInt interface{}) error {
	app := appInt.(*App)

	if tx.Qty <= math.Ndau(0) {
		return errors.New("invalid transfer: Qty not positive")
	}

	if tx.Source == tx.Destination {
		return errors.New("invalid transfer: source == destination")
	}

	source, _, _, err := app.getTxAccount(tx)
	if err != nil {
		return err
	}

	if source.IsLocked(app.blockTime) {
		return errors.New("source is locked")
	}

	dest, _ := app.getAccount(tx.Destination)

	if dest.IsNotified(app.blockTime) {
		return errors.New("transfers into notified addresses are invalid")
	}

	return nil
}

// Apply satisfies metatx.Transactable
func (tx *Transfer) Apply(appInt interface{}) error {
	app := appInt.(*App)
	err := app.applyTxDetails(tx)
	if err != nil {
		return err
	}

	source, _ := app.getAccount(tx.Source)
	dest, _ := app.getAccount(tx.Destination)

	err = (&dest.WeightedAverageAge).UpdateWeightedAverageAge(
		app.blockTime.Since(dest.LastWAAUpdate),
		tx.Qty,
		dest.Balance,
	)
	if err != nil {
		return errors.Wrap(err, "update waa")
	}
	dest.LastWAAUpdate = app.blockTime

	dest.Balance += tx.Qty
	if source.SettlementSettings.Period != 0 {
		dest.Settlements = append(dest.Settlements, backing.Settlement{
			Qty:    tx.Qty,
			Expiry: app.blockTime.Add(source.SettlementSettings.Period),
		})
	}

	dest.UpdateCurrencySeat(app.blockTime)

	return app.UpdateState(func(stateI metast.State) (metast.State, error) {
		state := stateI.(*backing.State)

		state.Accounts[tx.Destination.String()] = dest
		state.Accounts[tx.Source.String()] = source

		return state, nil
	})
}

// GetSource implements sourcer
func (tx *Transfer) GetSource(*App) (address.Address, error) {
	return tx.Source, nil
}

// Withdrawal implements withdrawer
func (tx *Transfer) Withdrawal() math.Ndau {
	return tx.Qty
}

// GetSequence implements sequencer
func (tx *Transfer) GetSequence() uint64 {
	return tx.Sequence
}

// GetSignatures implements signeder
func (tx *Transfer) GetSignatures() []signature.Signature {
	return tx.Signatures
}

// ExtendSignatures implements Signable
func (tx *Transfer) ExtendSignatures(sa []signature.Signature) {
	tx.Signatures = append(tx.Signatures, sa...)
}
