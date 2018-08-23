package ndau

import (
	"encoding/binary"
	"fmt"

	metast "github.com/oneiro-ndev/metanode/pkg/meta/state"
	"github.com/oneiro-ndev/ndau/pkg/ndau/backing"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	"github.com/oneiro-ndev/signature/pkg/signature"
	"github.com/pkg/errors"
)

// NewSetRewardsDestination creates a new SetRewardsDestination transaction
func NewSetRewardsDestination(account, destination address.Address, sequence uint64, keys []signature.PrivateKey) *SetRewardsDestination {
	c := &SetRewardsDestination{
		Source:      account,
		Destination: destination,
		Sequence:    sequence,
	}
	for _, key := range keys {
		c.Signatures = append(c.Signatures, key.Sign(c.SignableBytes()))
	}
	return c
}

// SignableBytes implements Transactable
func (c *SetRewardsDestination) SignableBytes() []byte {
	bytes := make([]byte, 8, 8+len(c.Source.String())+len(c.Destination.String()))
	binary.BigEndian.PutUint64(bytes[0:8], c.Sequence)
	bytes = append(bytes, c.Source.String()...)
	bytes = append(bytes, c.Destination.String()...)
	return bytes
}

// Validate implements metatx.Transactable
func (c *SetRewardsDestination) Validate(appI interface{}) error {
	app := appI.(*App)
	state := app.GetState().(*backing.State)

	accountData, hasAccount, _, err := state.GetValidAccount(
		c.Source,
		app.blockTime,
		c.Sequence,
		c.SignableBytes(),
		c.Signatures,
	)
	if err != nil {
		return err
	}
	if !hasAccount {
		return errors.New("No such account")
	}
	_, err = address.Validate(c.Destination.String())
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
	if c.Destination.String() != c.Source.String() {
		// dest account must not be sending rewards to any other account
		targetData, _ := state.GetAccount(c.Destination, app.blockTime)
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
func (c *SetRewardsDestination) Apply(appI interface{}) error {
	app := appI.(*App)

	return app.UpdateState(func(stateI metast.State) (metast.State, error) {
		state := stateI.(*backing.State)
		accountData, _ := state.GetAccount(c.Source, app.blockTime)
		accountData.Sequence = c.Sequence

		targetData, _ := state.GetAccount(c.Destination, app.blockTime)

		// update inbound of rewards target
		if accountData.RewardsTarget != nil && accountData.RewardsTarget.String() != c.Source.String() {
			oldTargetData, _ := state.GetAccount(*accountData.RewardsTarget, app.blockTime)
			// remove account from current target inbounds list
			inbounds := make([]address.Address, 0, len(oldTargetData.IncomingRewardsFrom)-1)
			for _, addr := range oldTargetData.IncomingRewardsFrom {
				if c.Source.String() != addr.String() {
					inbounds = append(inbounds, addr)
				}
			}
			oldTargetData.IncomingRewardsFrom = inbounds
			state.Accounts[accountData.RewardsTarget.String()] = oldTargetData
		}

		if c.Source.String() == c.Destination.String() {
			accountData.RewardsTarget = nil
		} else {
			accountData.RewardsTarget = &c.Destination
			targetData.IncomingRewardsFrom = append(targetData.IncomingRewardsFrom, c.Source)
			state.Accounts[c.Destination.String()] = targetData
		}

		state.Accounts[c.Source.String()] = accountData
		return state, nil
	})
}
