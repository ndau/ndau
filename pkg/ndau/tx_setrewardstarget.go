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

// NewSetRewardsTarget creates a new SetRewardsTarget transaction
func NewSetRewardsTarget(account, destination address.Address, sequence uint64, key signature.PrivateKey) *SetRewardsTarget {
	c := &SetRewardsTarget{
		Account:     account,
		Destination: destination,
		Sequence:    sequence,
	}
	c.Signature = key.Sign(c.SignableBytes())
	return c
}

// SignableBytes implements Transactable
func (c *SetRewardsTarget) SignableBytes() []byte {
	bytes := make([]byte, 8, 8+len(c.Account.String())+len(c.Destination.String()))
	binary.BigEndian.PutUint64(bytes[0:8], c.Sequence)
	bytes = append(bytes, c.Account.String()...)
	bytes = append(bytes, c.Destination.String()...)
	return bytes
}

// Validate implements metatx.Transactable
func (c *SetRewardsTarget) Validate(appI interface{}) error {
	app := appI.(*App)
	state := app.GetState().(*backing.State)

	accountData, hasAccount, err := state.GetValidAccount(
		c.Account,
		app.blockTime,
		c.Sequence,
		c.SignableBytes(),
		[]signature.Signature{c.Signature},
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
	if c.Destination.String() != c.Account.String() {
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
func (c *SetRewardsTarget) Apply(appI interface{}) error {
	app := appI.(*App)

	return app.UpdateState(func(stateI metast.State) (metast.State, error) {
		state := stateI.(*backing.State)
		accountData, _ := state.GetAccount(c.Account, app.blockTime)
		accountData.Sequence = c.Sequence

		targetData, _ := state.GetAccount(c.Destination, app.blockTime)

		// update inbound of rewards target
		if accountData.RewardsTarget != nil && accountData.RewardsTarget.String() != c.Account.String() {
			oldTargetData, _ := state.GetAccount(*accountData.RewardsTarget, app.blockTime)
			// remove account from current target inbounds list
			inbounds := make([]address.Address, 0, len(oldTargetData.IncomingRewardsFrom)-1)
			for _, addr := range oldTargetData.IncomingRewardsFrom {
				if c.Account.String() != addr.String() {
					inbounds = append(inbounds, addr)
				}
			}
			oldTargetData.IncomingRewardsFrom = inbounds
			state.Accounts[accountData.RewardsTarget.String()] = oldTargetData
		}

		if c.Account.String() == c.Destination.String() {
			accountData.RewardsTarget = nil
		} else {
			accountData.RewardsTarget = &c.Destination
			targetData.IncomingRewardsFrom = append(targetData.IncomingRewardsFrom, c.Account)
			state.Accounts[c.Destination.String()] = targetData
		}

		state.Accounts[c.Account.String()] = accountData
		return state, nil
	})
}
