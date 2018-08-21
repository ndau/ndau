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
func NewLock(account address.Address, period math.Duration, sequence uint64, keys []signature.PrivateKey) *Lock {
	c := &Lock{Target: account, Period: period, Sequence: sequence}
	for _, key := range keys {
		c.Signatures = append(c.Signatures, key.Sign(c.SignableBytes()))
	}
	return c
}

// SignableBytes implements Transactable
func (c *Lock) SignableBytes() []byte {
	bytes := make([]byte, 8+8, 8+8+len(c.Target.String()))
	binary.BigEndian.PutUint64(bytes[0:8], c.Sequence)
	binary.BigEndian.PutUint64(bytes[8:16], uint64(c.Period))
	bytes = append(bytes, c.Target.String()...)
	return bytes
}

// Validate implements metatx.Transactable
func (c *Lock) Validate(appI interface{}) error {
	app := appI.(*App)
	state := app.GetState().(*backing.State)

	accountData, hasAccount, err := state.GetValidAccount(
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
		accountData, _ := state.GetAccount(c.Target, app.blockTime)
		accountData.Sequence = c.Sequence

		accountData.Lock = &backing.Lock{
			NoticePeriod: c.Period,
		}

		state.Accounts[c.Target.String()] = accountData
		return state, nil
	})
}
