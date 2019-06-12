package ndau

import (
	"fmt"

	"github.com/oneiro-ndev/ndau/pkg/ndau/backing"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	"github.com/oneiro-ndev/ndaumath/pkg/signature"
	math "github.com/oneiro-ndev/ndaumath/pkg/types"
	"github.com/pkg/errors"
)

// Validate implements metatx.Transactable
func (tx *Unstake) Validate(appI interface{}) error {
	_, err := address.Validate(tx.Target.String())
	if err != nil {
		return errors.Wrap(err, "target")
	}
	_, err = address.Validate(tx.StakeTo.String())
	if err != nil {
		return errors.Wrap(err, "stake_to")
	}
	_, err = address.Validate(tx.Rules.String())
	if err != nil {
		return errors.Wrap(err, "rules")
	}

	app := appI.(*App)
	_, _, _, err = app.getTxAccount(tx)
	if err != nil {
		return errors.Wrap(err, "sequence")
	}

	vm, err := BuildVMForRulesValidation(tx, app.GetState().(*backing.State))
	if err != nil {
		return errors.Wrap(err, "building rules validation vm")
	}
	err = vm.Run(nil)
	if err != nil {
		return errors.Wrap(err, "running rules validation vm")
	}
	returncode, err := vm.Stack().PopAsInt64()
	if err != nil {
		return errors.Wrap(err, "getting return code from rules validation vm")
	}
	if returncode != 0 {
		return fmt.Errorf("rules validation script returned code %d", returncode)
	}

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

	// recalculate the validation rules. This time we're not interested in
	// the stack top, but its second value. If a second value is present,
	// it's a duration to retain the hold for from the block time.
	vm, err := BuildVMForRulesValidation(tx, app.GetState().(*backing.State))
	if err != nil {
		return errors.Wrap(err, "building rules validation vm")
	}
	err = vm.Run(nil)
	if err != nil {
		return errors.Wrap(err, "running rules validation vm")
	}
	// skip the stack top value; it was already validated
	stack := vm.Stack()
	_, err = stack.Pop()
	if err != nil {
		return errors.Wrap(err, "getting top stack value")
	}
	var retainFor math.Duration
	if stack.Depth() > 0 {
		retainI, err := vm.Stack().PopAsInt64()
		if err != nil {
			return errors.Wrap(err, "getting retain duration from rules vm")
		}
		retainFor = math.Duration(retainI)
	}

	return app.UpdateState(app.Unstake(tx.Qty, tx.Target, tx.StakeTo, tx.Rules, retainFor))
}

// GetSource implements Sourcer
func (tx *Unstake) GetSource(*App) (address.Address, error) {
	return tx.Target, nil
}

// GetSequence implements Sequencer
func (tx *Unstake) GetSequence() uint64 {
	return tx.Sequence
}

// GetSignatures implements Signeder
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
