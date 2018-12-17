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
func (tx *ChangeSettlementPeriod) GetAccountAddresses() []string {
	return []string{tx.Target.String()}
}

// NewChangeSettlementPeriod creates a new signed settlement period change
func NewChangeSettlementPeriod(
	target address.Address,
	newPeriod math.Duration,
	sequence uint64,
	keys []signature.PrivateKey,
) (ChangeSettlementPeriod, error) {
	tx := ChangeSettlementPeriod{
		Target:   target,
		Period:   newPeriod,
		Sequence: sequence,
	}
	sb := tx.SignableBytes()
	for _, key := range keys {
		tx.Signatures = append(tx.Signatures, key.Sign(sb))
	}
	return tx, nil
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
		acct, _ := state.GetAccount(tx.Target, app.blockTime)

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
