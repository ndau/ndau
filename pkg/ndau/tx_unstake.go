package ndau

import (
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	"github.com/oneiro-ndev/ndaumath/pkg/signature"
	"github.com/pkg/errors"
)

// Validate implements metatx.Transactable
func (tx *Unstake) Validate(appI interface{}) error {
	_, err := address.Validate(tx.Target.String())
	if err != nil {
		return errors.Wrap(err, "target")
	}

	app := appI.(*App)
	_, _, _, err = app.getTxAccount(tx)
	return err
}

// Apply implements metatx.Transactable
func (tx *Unstake) Apply(appI interface{}) error {
	app := appI.(*App)

	var err error
	err = app.applyTxDetails(tx)
	if err != nil {
		return err
	}
	return app.UpdateState(app.Unstake(tx.Qty, tx.Target, tx.StakeTo, tx.Rules))
}

// GetSource implements sourcer
func (tx *Unstake) GetSource(*App) (address.Address, error) {
	return tx.Target, nil
}

// GetSequence implements sequencer
func (tx *Unstake) GetSequence() uint64 {
	return tx.Sequence
}

// GetSignatures implements signeder
func (tx *Unstake) GetSignatures() []signature.Signature {
	return tx.Signatures
}

// ExtendSignatures implements Signable
func (tx *Unstake) ExtendSignatures(sa []signature.Signature) {
	tx.Signatures = append(tx.Signatures, sa...)
}

// GetAccountAddresses returns the account addresses associated with this transaction type.
func (tx *Unstake) GetAccountAddresses() []string {
	return []string{tx.Target.String()}
}
