package ndau

import (
	metast "github.com/oneiro-ndev/metanode/pkg/meta/state"
	"github.com/oneiro-ndev/ndau/pkg/ndau/backing"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	"github.com/oneiro-ndev/ndaumath/pkg/signature"
	"github.com/pkg/errors"
)

// GetAccountAddresses returns the account addresses associated with this transaction type.
func (tx *Notify) GetAccountAddresses() []string {
	return []string{tx.Target.String()}
}

// Validate implements metatx.Transactable
func (tx *Notify) Validate(appI interface{}) error {
	app := appI.(*App)

	accountData, hasAccount, _, err := app.getTxAccount(tx)
	if err != nil {
		return err
	}

	if !hasAccount {
		return errors.New("No such account")
	}

	if accountData.Lock == nil {
		return errors.New("Account is not locked")
	}
	if accountData.Lock.UnlocksOn != nil {
		return errors.New("Account has already been notified")
	}

	return nil
}

// Apply implements metatx.Transactable
func (tx *Notify) Apply(appI interface{}) error {
	app := appI.(*App)
	err := app.applyTxDetails(tx)
	if err != nil {
		return err
	}

	return app.UpdateState(func(stateI metast.State) (metast.State, error) {
		state := stateI.(*backing.State)
		accountData, _ := app.getAccount(tx.Target)

		uo := app.BlockTime().Add(accountData.Lock.NoticePeriod)
		accountData.Lock.UnlocksOn = &uo

		state.Accounts[tx.Target.String()] = accountData
		return state, nil
	})
}

// GetSource implements Sourcer
func (tx *Notify) GetSource(*App) (address.Address, error) {
	return tx.Target, nil
}

// GetSequence implements Sequencer
func (tx *Notify) GetSequence() uint64 {
	return tx.Sequence
}

// GetSignatures implements Signeder
func (tx *Notify) GetSignatures() []signature.Signature {
	return tx.Signatures
}

// ExtendSignatures implements Signable
func (tx *Notify) ExtendSignatures(sa []signature.Signature) {
	tx.Signatures = append(tx.Signatures, sa...)
}
