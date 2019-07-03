package ndau

import (
	"fmt"

	"github.com/oneiro-ndev/ndau/pkg/ndau/backing"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	"github.com/oneiro-ndev/ndaumath/pkg/signature"
	math "github.com/oneiro-ndev/ndaumath/pkg/types"
	sv "github.com/oneiro-ndev/system_vars/pkg/system_vars"
	"github.com/pkg/errors"
)

// GetAccountAddresses returns the account addresses associated with this transaction type.
func (tx *Stake) GetAccountAddresses() []string {
	return []string{tx.Target.String(), tx.StakeTo.String(), tx.Rules.String()}
}

// Validate implements metatx.Transactable
func (tx *Stake) Validate(appI interface{}) error {
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
	target, hasAccount, _, err := app.getTxAccount(tx)
	if err != nil {
		return err
	}
	if !hasAccount {
		return errors.New("target does not exist")
	}

	if tx.StakeTo == tx.Rules {
		ps := target.PrimaryStake(tx.Rules)
		if ps != nil {
			return fmt.Errorf("stake: cannot have more than 1 primary stake to a rules account")
		}
	}

	var minStake math.Ndau
	err = app.System(sv.MinNodeRegistrationStakeName, &minStake)
	if err != nil {
		return errors.Wrap(err, "fetching MinStake system variable")
	}

	if tx.Qty < minStake {
		return fmt.Errorf("cannot stake %s ndau: must stake at least MinStake (%s)", tx.Qty, minStake)
	}

	txFee, err := app.calculateTxFee(tx)
	if err != nil {
		return errors.Wrap(err, "calculating tx fee")
	}

	requiredBalance, err := minStake.Add(txFee)
	if err != nil {
		return errors.Wrap(err, "calculating required balance")
	}

	if target.Balance.Compare(requiredBalance) < 0 {
		return fmt.Errorf("target has insufficient balance: have %s ndau, need %s", target.Balance, minStake)
	}

	_, hasNode := app.getAccount(tx.StakeTo)
	if !hasNode {
		return errors.New("Node does not exist")
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

	return nil
}

// Apply implements metatx.Transactable
func (tx *Stake) Apply(appI interface{}) error {
	app := appI.(*App)
	return app.UpdateState(
		app.applyTxDetails(tx),
		app.Stake(tx.Qty, tx.Target, tx.StakeTo, tx.Rules, tx))
}

// GetSource implements Sourcer
func (tx *Stake) GetSource(*App) (address.Address, error) {
	return tx.Target, nil
}

// GetSequence implements Sequencer
func (tx *Stake) GetSequence() uint64 {
	return tx.Sequence
}

// GetSignatures implements Signeder
func (tx *Stake) GetSignatures() []signature.Signature {
	return tx.Signatures
}

// ExtendSignatures implements Signable
func (tx *Stake) ExtendSignatures(sa []signature.Signature) {
	tx.Signatures = append(tx.Signatures, sa...)
}
