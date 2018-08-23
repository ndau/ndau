package ndau

import (
	"encoding/binary"

	metast "github.com/oneiro-ndev/metanode/pkg/meta/state"
	"github.com/oneiro-ndev/ndau/pkg/ndau/backing"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	"github.com/oneiro-ndev/signature/pkg/signature"
	"github.com/pkg/errors"
)

// NewNotify creates a new Notify transaction
func NewNotify(account address.Address, sequence uint64, keys []signature.PrivateKey) *Notify {
	c := &Notify{Target: account, Sequence: sequence}
	for _, key := range keys {
		c.Signatures = append(c.Signatures, key.Sign(c.SignableBytes()))
	}
	return c
}

// SignableBytes implements Transactable
func (c *Notify) SignableBytes() []byte {
	bytes := make([]byte, 8, 8+len(c.Target.String()))
	binary.BigEndian.PutUint64(bytes[0:8], c.Sequence)
	bytes = append(bytes, c.Target.String()...)
	return bytes
}

// Validate implements metatx.Transactable
func (c *Notify) Validate(appI interface{}) error {
	app := appI.(*App)
	state := app.GetState().(*backing.State)

	accountData, hasAccount, _, err := state.GetValidAccount(
		c.Target,
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

	if accountData.Lock == nil {
		return errors.New("Account is not locked")
	}
	if accountData.Lock.UnlocksOn != nil {
		return errors.New("Account has already been notified")
	}

	return nil
}

// Apply implements metatx.Transactable
func (c *Notify) Apply(appI interface{}) error {
	app := appI.(*App)

	return app.UpdateState(func(stateI metast.State) (metast.State, error) {
		state := stateI.(*backing.State)
		accountData, _ := state.GetAccount(c.Target, app.blockTime)
		accountData.Sequence = c.Sequence
		uo := app.blockTime.Add(accountData.Lock.NoticePeriod)
		accountData.Lock.UnlocksOn = &uo

		state.Accounts[c.Target.String()] = accountData
		return state, nil
	})
}
