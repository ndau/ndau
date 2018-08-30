package ndau

import (
	metast "github.com/oneiro-ndev/metanode/pkg/meta/state"
	"github.com/oneiro-ndev/ndau/pkg/ndau/backing"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	math "github.com/oneiro-ndev/ndaumath/pkg/types"
	"github.com/oneiro-ndev/signature/pkg/signature"
	"github.com/pkg/errors"
)

// NewLock creates a new Lock transaction
func NewLock(account address.Address, period math.Duration, sequence uint64, keys []signature.PrivateKey) *Lock {
	tx := &Lock{Target: account, Period: period, Sequence: sequence}
	for _, key := range keys {
		tx.Signatures = append(tx.Signatures, key.Sign(tx.SignableBytes()))
	}
	return tx
}

// SignableBytes implements Transactable
func (tx *Lock) SignableBytes() []byte {
	bytes := make([]byte, 0, 8+8+len(tx.Target.String()))
	bytes = appendUint64(bytes, tx.Sequence)
	bytes = appendUint64(bytes, uint64(tx.Period))
	bytes = append(bytes, tx.Target.String()...)
	return bytes
}

// Validate implements metatx.Transactable
func (tx *Lock) Validate(appI interface{}) error {
	app := appI.(*App)

	accountData, hasAccount, _, err := app.getTxAccount(
		tx,
		tx.Target,
		tx.Sequence,
		tx.Signatures,
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
			if tx.Period < accountData.Lock.NoticePeriod {
				return errors.New("Locked, non-notified accounts may be relocked only with periods >= their current")
			}
		} else {
			// if notified, lock is valid if it expires after the current unlock time
			lockExpiry := app.blockTime.Add(tx.Period)
			uo := *accountData.Lock.UnlocksOn
			if lockExpiry.Compare(uo) < 0 {
				return errors.New("Locked, notified accounts may be relocked only when new lock min expiry >= current unlock time")
			}
		}
	}

	return nil
}

// Apply implements metatx.Transactable
func (tx *Lock) Apply(appI interface{}) error {
	app := appI.(*App)

	return app.UpdateState(func(stateI metast.State) (metast.State, error) {
		state := stateI.(*backing.State)
		accountData, _ := state.GetAccount(tx.Target, app.blockTime)
		accountData.Sequence = tx.Sequence

		accountData.Lock = &backing.Lock{
			NoticePeriod: tx.Period,
		}

		state.Accounts[tx.Target.String()] = accountData
		return state, nil
	})
}

// CalculateTxFee implements metatx.Transactable
func (tx *Lock) CalculateTxFee(appI interface{}) (math.Ndau, error) {
	app := appI.(*App)
	return app.calculateTxFee(tx)
}
