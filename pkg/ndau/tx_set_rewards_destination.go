package ndau

import (
	"fmt"

	metast "github.com/oneiro-ndev/metanode/pkg/meta/state"
	"github.com/oneiro-ndev/ndau/pkg/ndau/backing"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	"github.com/oneiro-ndev/signature/pkg/signature"
	"github.com/pkg/errors"
)

// NewSetRewardsDestination creates a new SetRewardsDestination transaction
func NewSetRewardsDestination(account, destination address.Address, sequence uint64, keys []signature.PrivateKey) *SetRewardsDestination {
	tx := &SetRewardsDestination{
		Source:      account,
		Destination: destination,
		Sequence:    sequence,
	}
	for _, key := range keys {
		tx.Signatures = append(tx.Signatures, key.Sign(tx.SignableBytes()))
	}
	return tx
}

// SignableBytes implements Transactable
func (tx *SetRewardsDestination) SignableBytes() []byte {
	bytes := make([]byte, 0, 8+len(tx.Source.String())+len(tx.Destination.String()))
	bytes = appendUint64(bytes, tx.Sequence)
	bytes = append(bytes, tx.Source.String()...)
	bytes = append(bytes, tx.Destination.String()...)
	return bytes
}

// Validate implements metatx.Transactable
func (tx *SetRewardsDestination) Validate(appI interface{}) error {
	app := appI.(*App)
	state := app.GetState().(*backing.State)

	accountData, hasAccount, _, err := app.getTxAccount(
		tx,
		tx.Source,
		tx.Sequence,
		tx.Signatures,
	)
	if err != nil {
		return err
	}
	if !hasAccount {
		return errors.New("No such account")
	}
	_, err = address.Validate(tx.Destination.String())
	if err != nil {
		return errors.Wrap(err, "Destination")
	}

	// source account must not be receiving rewards from any other account
	if len(accountData.IncomingRewardsFrom) > 0 {
		return fmt.Errorf("Accounts may not both send and receive rewards. Source receives rewards from these accounts: %s", accountData.IncomingRewardsFrom)
	}

	// if dest is the same as source, we're resetting the EAI to accumulate
	// in its account of origin.
	// neither destination rule appllies in that case.
	if tx.Destination.String() != tx.Source.String() {
		// dest account must not be sending rewards to any other account
		targetData, _ := state.GetAccount(tx.Destination, app.blockTime)
		if targetData.RewardsTarget != nil {
			return fmt.Errorf("Accounts may not both send and receive rewards. Destination sends rewards to %s", *targetData.RewardsTarget)
		}

		// dest account must not be notified
		if targetData.IsNotified(app.blockTime) {
			return errors.New("Destination is currently notified and may not receive new rewards until it unlocks")
		}
	}

	return nil
}

// Apply implements metatx.Transactable
func (tx *SetRewardsDestination) Apply(appI interface{}) error {
	app := appI.(*App)

	return app.UpdateState(func(stateI metast.State) (metast.State, error) {
		state := stateI.(*backing.State)
		accountData, _ := state.GetAccount(tx.Source, app.blockTime)
		accountData.Sequence = tx.Sequence

		targetData, _ := state.GetAccount(tx.Destination, app.blockTime)

		// update inbound of rewards target
		if accountData.RewardsTarget != nil && accountData.RewardsTarget.String() != tx.Source.String() {
			oldTargetData, _ := state.GetAccount(*accountData.RewardsTarget, app.blockTime)
			// remove account from current target inbounds list
			inbounds := make([]address.Address, 0, len(oldTargetData.IncomingRewardsFrom)-1)
			for _, addr := range oldTargetData.IncomingRewardsFrom {
				if tx.Source.String() != addr.String() {
					inbounds = append(inbounds, addr)
				}
			}
			oldTargetData.IncomingRewardsFrom = inbounds
			state.Accounts[accountData.RewardsTarget.String()] = oldTargetData
		}

		if tx.Source.String() == tx.Destination.String() {
			accountData.RewardsTarget = nil
		} else {
			accountData.RewardsTarget = &tx.Destination
			targetData.IncomingRewardsFrom = append(targetData.IncomingRewardsFrom, tx.Source)
			state.Accounts[tx.Destination.String()] = targetData
		}

		state.Accounts[tx.Source.String()] = accountData
		return state, nil
	})
}
