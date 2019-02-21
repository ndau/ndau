package ndau

import (
	metast "github.com/oneiro-ndev/metanode/pkg/meta/state"
	"github.com/oneiro-ndev/ndau/pkg/ndau/backing"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	"github.com/oneiro-ndev/ndaumath/pkg/eai"
	"github.com/oneiro-ndev/ndaumath/pkg/signature"
	math "github.com/oneiro-ndev/ndaumath/pkg/types"
	sv "github.com/oneiro-ndev/system_vars/pkg/system_vars"
	"github.com/pkg/errors"
)

// GetAccountAddresses returns the account addresses associated with this transaction type.
func (tx *TransferAndLock) GetAccountAddresses() []string {
	return []string{tx.Source.String(), tx.Destination.String()}
}

// Validate satisfies metatx.Transactable
func (tx *TransferAndLock) Validate(appInt interface{}) error {
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

	_, exists := app.getAccount(tx.Destination)

	// TransferAndLock cannot be sent to an existing account
	if exists {
		return errors.New("invalid TransferAndLock: cannot be sent to an existing account")
	}

	return nil
}

// Apply satisfies metatx.Transactable
func (tx *TransferAndLock) Apply(appInt interface{}) error {
	app := appInt.(*App)
	err := app.applyTxDetails(tx)
	if err != nil {
		return err
	}

	lockedBonusRateTable := eai.RateTable{}
	err = app.System(sv.LockedRateTableName, &lockedBonusRateTable)
	if err != nil {
		return err
	}

	source, _ := app.getAccount(tx.Source)
	// we know dest is a new account so WAA, WAA update and EAI update times are set properly
	dest, _ := app.getAccount(tx.Destination)
	dest.Balance = tx.Qty
	if source.SettlementSettings.Period != 0 {
		dest.Settlements = append(dest.Settlements, backing.Settlement{
			Qty:    tx.Qty,
			Expiry: app.blockTime.Add(source.SettlementSettings.Period),
		})
	}
	dest.Lock = backing.NewLock(tx.Period, lockedBonusRateTable)

	dest.UpdateCurrencySeat(app.blockTime)

	return app.UpdateState(func(stateI metast.State) (metast.State, error) {
		state := stateI.(*backing.State)

		state.Accounts[tx.Destination.String()] = dest
		state.Accounts[tx.Source.String()] = source

		return state, nil
	})
}

// GetSource implements sourcer
func (tx *TransferAndLock) GetSource(*App) (address.Address, error) {
	return tx.Source, nil
}

// Withdrawal implements withdrawer
func (tx *TransferAndLock) Withdrawal() math.Ndau {
	return tx.Qty
}

// GetSequence implements sequencer
func (tx *TransferAndLock) GetSequence() uint64 {
	return tx.Sequence
}

// GetSignatures implements signeder
func (tx *TransferAndLock) GetSignatures() []signature.Signature {
	return tx.Signatures
}

// ExtendSignatures implements Signable
func (tx *TransferAndLock) ExtendSignatures(sa []signature.Signature) {
	tx.Signatures = append(tx.Signatures, sa...)
}
