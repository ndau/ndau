package ndau

import (
	"fmt"

	metast "github.com/oneiro-ndev/metanode/pkg/meta/state"
	"github.com/oneiro-ndev/ndau/pkg/ndau/backing"
	sv "github.com/oneiro-ndev/ndau/pkg/ndau/system_vars"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	math "github.com/oneiro-ndev/ndaumath/pkg/types"
	"github.com/oneiro-ndev/signature/pkg/signature"
	"github.com/pkg/errors"
)

// NewStake creates a new signed Stake transaction
func NewStake(
	target, node address.Address,
	sequence uint64,
	keys []signature.PrivateKey,
) *Stake {
	tx := &Stake{
		Target:   target,
		Node:     node,
		Sequence: sequence,
	}
	for _, key := range keys {
		tx.Signatures = append(tx.Signatures, key.Sign(tx.SignableBytes()))
	}
	return tx
}

// SignableBytes implements Transactable
func (tx *Stake) SignableBytes() []byte {
	bytes := make([]byte, 0, tx.Target.Msgsize()+tx.Node.Msgsize()+8)
	bytes = appendUint64(bytes, tx.Sequence)
	bytes = append(bytes, []byte(tx.Target.String())...)
	bytes = append(bytes, []byte(tx.Node.String())...)
	return bytes
}

// Validate implements metatx.Transactable
func (tx *Stake) Validate(appI interface{}) error {
	_, err := address.Validate(tx.Target.String())
	if err != nil {
		return errors.Wrap(err, "target")
	}
	_, err = address.Validate(tx.Node.String())
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
	err = app.System(sv.MinStakeName, &minStake)
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

	node, hasNode := app.GetState().(*backing.State).GetAccount(tx.Node, app.blockTime)
	if !hasNode {
		return errors.New("Node does not exist")
	}
	if node.Stake == nil || node.Stake.Address != tx.Node {
		return errors.New("Node is not self-staked")
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
			Address: tx.Node,
			Point:   app.blockTime,
		}

		state.Accounts[tx.Target.String()] = target
		err = state.Stake(tx.Target, tx.Node)

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
