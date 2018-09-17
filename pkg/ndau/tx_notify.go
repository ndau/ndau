package ndau

import (
	metast "github.com/oneiro-ndev/metanode/pkg/meta/state"
	"github.com/oneiro-ndev/ndau/pkg/ndau/backing"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
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

	accountData, hasAccount, _, err := app.getTxAccount(tx)
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
	err := app.applyTxDetails(tx)
	if err != nil {
		return err
	}

	return app.UpdateState(func(stateI metast.State) (metast.State, error) {
		state := stateI.(*backing.State)
		accountData, _ := state.GetAccount(tx.Target, app.blockTime)

		uo := app.blockTime.Add(accountData.Lock.NoticePeriod)
		accountData.Lock.UnlocksOn = &uo

		state.Accounts[tx.Target.String()] = accountData
		return state, nil
	})
}

// GetSource implements sourcer
func (tx *Notify) GetSource(*App) (address.Address, error) {
	return tx.Target, nil
}

// GetSequence implements sequencer
func (tx *Notify) GetSequence() uint64 {
	return tx.Sequence
}

// GetSignatures implements signeder
func (tx *Notify) GetSignatures() []signature.Signature {
	return tx.Signatures
}
