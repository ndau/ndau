package ndau

import (
	"encoding/binary"

	metast "github.com/oneiro-ndev/metanode/pkg/meta/state"
	"github.com/oneiro-ndev/ndau/pkg/ndau/backing"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	"github.com/oneiro-ndev/signature/pkg/signature"
	"github.com/pkg/errors"
)

// NewSetRewardsTarget creates a new SetRewardsTarget transaction
//
// Most users will never need this.
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

	accountData, hasAccount := state.GetAccount(c.Account, app.blockTime)
	if !hasAccount {
		return errors.New("No such account")
	}
	// is the tx sequence higher than the highest previous sequence?
	if c.Sequence <= accountData.Sequence {
		return errors.New("Sequence too low")
	}
	// does the signature check out?
	if accountData.TransferKey == nil {
		return errors.New("Transfer key not set")
	}
	if !accountData.TransferKey.Verify(c.SignableBytes(), c.Signature) {
		return errors.New("Invalid signature")
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

		accountData.RewardsTarget = &c.Destination

		state.Accounts[c.Account.String()] = accountData
		return state, nil
	})
}
