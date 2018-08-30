package ndau

import (
	metast "github.com/oneiro-ndev/metanode/pkg/meta/state"
	"github.com/oneiro-ndev/ndau/pkg/ndau/backing"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	math "github.com/oneiro-ndev/ndaumath/pkg/types"
	"github.com/oneiro-ndev/signature/pkg/signature"
	"github.com/pkg/errors"
)

// NewNotify creates a new Notify transaction
func NewNotify(account address.Address, sequence uint64, keys []signature.PrivateKey) *Notify {
	tx := &Notify{Target: account, Sequence: sequence}
	for _, key := range keys {
		tx.Signatures = append(tx.Signatures, key.Sign(tx.SignableBytes()))
	}
	return tx
}

// SignableBytes implements Transactable
func (tx *Notify) SignableBytes() []byte {
	bytes := make([]byte, 0, 8+len(tx.Target.String()))
	bytes = appendUint64(bytes, tx.Sequence)
	bytes = append(bytes, tx.Target.String()...)
	return bytes
}

// Validate implements metatx.Transactable
func (tx *Notify) Validate(appI interface{}) error {
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

	if accountData.Lock == nil {
		return errors.New("Account is not locked")
	}
	if accountData.Lock.UnlocksOn != nil {
		return errors.New("Account has already been notified")
	}

	return nil
}

// Apply implements metatx.Transactable
func (tx *Notify) Apply(appI interface{}) error {
	app := appI.(*App)

	return app.UpdateState(func(stateI metast.State) (metast.State, error) {
		state := stateI.(*backing.State)
		accountData, _ := state.GetAccount(tx.Target, app.blockTime)
		accountData.Sequence = tx.Sequence
		uo := app.blockTime.Add(accountData.Lock.NoticePeriod)
		accountData.Lock.UnlocksOn = &uo

		state.Accounts[tx.Target.String()] = accountData
		return state, nil
	})
}

// CalculateTxFee implements metatx.Transactable
func (tx *Notify) CalculateTxFee(appI interface{}) (math.Ndau, error) {
	app := appI.(*App)
	return app.calculateTxFee(tx)
}
