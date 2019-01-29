package ndau

import (
	"fmt"

	metast "github.com/oneiro-ndev/metanode/pkg/meta/state"
	"github.com/oneiro-ndev/ndau/pkg/ndau/backing"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	"github.com/oneiro-ndev/ndaumath/pkg/signature"
	math "github.com/oneiro-ndev/ndaumath/pkg/types"
	sv "github.com/oneiro-ndev/system_vars/pkg/system_vars"
	"github.com/pkg/errors"
)

// GetAccountAddresses returns the account addresses associated with this transaction type.
func (tx *Stake) GetAccountAddresses() []string {
	return []string{tx.Target.String(), tx.StakedAccount.String()}
}

// Validate implements metatx.Transactable
func (tx *Stake) Validate(appI interface{}) error {
	_, err := address.Validate(tx.Target.String())
	if err != nil {
		return errors.Wrap(err, "target")
	}
	_, err = address.Validate(tx.StakedAccount.String())
	if err != nil {
		return errors.Wrap(err, "node")
	}

	app := appI.(*App)
	target, hasAccount, _, err := app.getTxAccount(tx)
	if err != nil {
		return err
	}
	if !hasAccount {
		return errors.New("target does not exist")
	}

	if target.Stake != nil {
		return errors.New("target must unstake and cooldown before re-staking")
	}

	var minStake math.Ndau
	err = app.System(sv.MinNodeRegistrationStakeName, &minStake)
	if err != nil {
		return errors.Wrap(err, "fetching MinStake system variable")
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

	node, hasNode := app.GetState().(*backing.State).GetAccount(tx.StakedAccount, app.blockTime)
	if !hasNode {
		return errors.New("Node does not exist")
	}
	if tx.StakedAccount != tx.Target {
		if node.Stake == nil || node.Stake.Address != tx.StakedAccount {
			return errors.New("Node is not self-staked")
		}
	}

	if target.Balance.Compare(node.Balance) > 0 {
		return fmt.Errorf("target balance (%s) may not exceed node balance (%s)", target.Balance, node.Balance)
	}

	return nil
}

// Apply implements metatx.Transactable
func (tx *Stake) Apply(appI interface{}) error {
	app := appI.(*App)

	return app.UpdateState(func(stateI metast.State) (metast.State, error) {
		state := stateI.(*backing.State)
		target, _ := state.GetAccount(tx.Target, app.blockTime)
		target.Sequence = tx.Sequence

		fee, err := app.calculateTxFee(tx)
		if err != nil {
			return state, err
		}
		target.Balance -= fee

		target.Stake = &backing.Stake{
			Address: tx.StakedAccount,
			Point:   app.blockTime,
		}

		state.Accounts[tx.Target.String()] = target
		err = state.Stake(tx.Target, tx.StakedAccount)

		return state, nil
	})
}

// GetSource implements sourcer
func (tx *Stake) GetSource(*App) (address.Address, error) {
	return tx.Target, nil
}

// GetSequence implements sequencer
func (tx *Stake) GetSequence() uint64 {
	return tx.Sequence
}

// GetSignatures implements signeder
func (tx *Stake) GetSignatures() []signature.Signature {
	return tx.Signatures
}

// ExtendSignatures implements Signable
func (tx *Stake) ExtendSignatures(sa []signature.Signature) {
	tx.Signatures = append(tx.Signatures, sa...)
}
