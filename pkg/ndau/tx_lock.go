package ndau

import (
	"encoding/binary"

	metast "github.com/oneiro-ndev/metanode/pkg/meta/state"
	"github.com/oneiro-ndev/ndau/pkg/ndau/backing"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	math "github.com/oneiro-ndev/ndaumath/pkg/types"
	"github.com/oneiro-ndev/signature/pkg/signature"
	"github.com/pkg/errors"
)

// NewLock creates a new Lock transaction
func NewLock(account address.Address, period math.Duration, sequence uint64, key signature.PrivateKey) *Lock {
	c := &Lock{Account: account, Period: period, Sequence: sequence}
	c.Signature = key.Sign(c.SignableBytes())
	return c
}

// SignableBytes implements Transactable
func (c *Lock) SignableBytes() []byte {
	bytes := make([]byte, 8+8, 8+8+len(c.Account.String()))
	binary.BigEndian.PutUint64(bytes[0:8], c.Sequence)
	binary.BigEndian.PutUint64(bytes[8:16], uint64(c.Period))
	bytes = append(bytes, c.Account.String()...)
	return bytes
}

// Validate implements metatx.Transactable
func (c *Lock) Validate(appI interface{}) error {
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

	if accountData.Lock != nil {
		if accountData.Lock.UnlocksOn == nil {
			// if not notified, lock is valid if its period >= the current period
			if c.Period < accountData.Lock.NoticePeriod {
				return errors.New("Locked, non-notified accounts may be relocked only with periods >= their current")
			}
		} else {
			// if notified, lock is valid if it expires after the current unlock time
			lockExpiry := app.blockTime.Add(c.Period)
			uo := *accountData.Lock.UnlocksOn
			if lockExpiry.Compare(uo) < 0 {
				return errors.New("Locked, notified accounts may be relocked only when new lock min expiry >= current unlock time")
			}
		}
	}

	return nil
}

// Apply implements metatx.Transactable
func (c *Lock) Apply(appI interface{}) error {
	app := appI.(*App)

	return app.UpdateState(func(stateI metast.State) (metast.State, error) {
		state := stateI.(*backing.State)
		accountData, _ := state.GetAccount(c.Account, app.blockTime)
		accountData.Sequence = c.Sequence

		accountData.Lock = &backing.Lock{
			NoticePeriod: c.Period,
		}

		state.Accounts[c.Account.String()] = accountData
		return state, nil
	})
}