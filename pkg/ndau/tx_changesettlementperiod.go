package ndau

import (
	metast "github.com/oneiro-ndev/metanode/pkg/meta/state"
	"github.com/oneiro-ndev/ndau/pkg/ndau/backing"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	"github.com/oneiro-ndev/ndaumath/pkg/signature"
	"github.com/pkg/errors"
)

// GetAccountAddresses returns the account addresses associated with this transaction type.
func (tx *ChangeSettlementPeriod) GetAccountAddresses() []string {
	return []string{tx.Target.String()}
}

// Validate implements metatx.Transactable
func (tx *ChangeSettlementPeriod) Validate(appI interface{}) (err error) {
	app := appI.(*App)

	if tx.Period < 0 {
		return errors.New("Negative settlement period")
	}
	_, _, _, err = app.getTxAccount(tx)
	if err != nil {
		return err
	}

	return nil
}

// Apply implements metatx.Transactable
func (tx *ChangeSettlementPeriod) Apply(appI interface{}) error {
	app := appI.(*App)
	err := app.applyTxDetails(tx)
	if err != nil {
		return err
	}

	return app.UpdateState(func(stateI metast.State) (metast.State, error) {
		state := stateI.(*backing.State)
		acct, _ := app.getAccount(tx.Target)

		ca := app.blockTime.Add(acct.SettlementSettings.Period)
		acct.SettlementSettings.ChangesAt = &ca
		acct.SettlementSettings.Next = &tx.Period

		state.Accounts[tx.Target.String()] = acct
		return state, nil
	})
}

// GetSource implements sourcer
func (tx *ChangeSettlementPeriod) GetSource(*App) (address.Address, error) {
	return tx.Target, nil
}

// GetSequence implements sequencer
func (tx *ChangeSettlementPeriod) GetSequence() uint64 {
	return tx.Sequence
}

// GetSignatures implements signeder
func (tx *ChangeSettlementPeriod) GetSignatures() []signature.Signature {
	return tx.Signatures
}

// ExtendSignatures implements Signable
func (tx *ChangeSettlementPeriod) ExtendSignatures(sa []signature.Signature) {
	tx.Signatures = append(tx.Signatures, sa...)
}
