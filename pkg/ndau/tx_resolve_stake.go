package ndau

import (
	"fmt"

	"github.com/oneiro-ndev/ndaumath/pkg/address"
	"github.com/oneiro-ndev/ndaumath/pkg/signature"
	"github.com/pkg/errors"
)

const resolveStakeDenominator = 255

// GetAccountAddresses returns the account addresses associated with this transaction type.
func (tx *ResolveStake) GetAccountAddresses() []string {
	return []string{tx.Target.String(), tx.Rules.String()}
}

// Validate implements metatx.Transactable
func (tx *ResolveStake) Validate(appI interface{}) error {
	_, err := address.Validate(tx.Target.String())
	if err != nil {
		return errors.Wrap(err, "target")
	}
	_, err = address.Validate(tx.Rules.String())
	if err != nil {
		return errors.Wrap(err, "rules")
	}

	// business rule: this field is interpreted as fraction
	// with implied denominator resolveStakeDenominator, indicating the disposition of
	// that portion of the staked amount.
	//
	// Note to future maintainers: when the value is less than 1, the remainder
	// of the stake is returned immediately to the stakers as spendable.
	// There is no cooldown because, once a human gets involved to issue
	// this tx, the stake is resolved and there is nothing further to do about it.
	if uint(tx.Burn) > resolveStakeDenominator {
		return fmt.Errorf("Burn must be <= %d", resolveStakeDenominator)
	}

	app := appI.(*App)

	rules, _, _, err := app.getTxAccount(tx)
	if err != nil {
		return err
	}
	if rules.StakeRules == nil {
		return fmt.Errorf("%s is not a rules account", tx.Rules)
	}

	target, _ := app.getAccount(tx.Target)

	if h := target.PrimaryStake(tx.Rules); h == nil {
		return fmt.Errorf("%s is not a primary staker to %s", tx.Target, tx.Rules)
	}

	return nil
}

// Apply implements metatx.Transactable
func (tx *ResolveStake) Apply(appI interface{}) error {
	app := appI.(*App)

	var err error
	err = app.applyTxDetails(tx)
	if err != nil {
		return err
	}

	// all state changes get applied or none do
	// UnstakeAndBurn ignores
	return app.UpdateState(app.UnstakeAndBurn(0, tx.Burn, tx.Target, tx.Target, tx.Rules, 0, true))
}

// GetSource implements Sourcer
func (tx *ResolveStake) GetSource(*App) (address.Address, error) {
	return tx.Rules, nil
}

// GetSequence implements Sequencer
func (tx *ResolveStake) GetSequence() uint64 {
	return tx.Sequence
}

// GetSignatures implements Signeder
func (tx *ResolveStake) GetSignatures() []signature.Signature {
	return tx.Signatures
}

// ExtendSignatures implements Signable
func (tx *ResolveStake) ExtendSignatures(sa []signature.Signature) {
	tx.Signatures = append(tx.Signatures, sa...)
}
