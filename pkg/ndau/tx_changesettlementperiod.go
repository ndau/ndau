package ndau

import (
	metast "github.com/oneiro-ndev/metanode/pkg/meta/state"
	"github.com/oneiro-ndev/ndau/pkg/ndau/backing"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	math "github.com/oneiro-ndev/ndaumath/pkg/types"
	"github.com/oneiro-ndev/signature/pkg/signature"
	"github.com/pkg/errors"
)

// SignableBytes implements Transactable
func (tx *ChangeSettlementPeriod) SignableBytes() []byte {
	bytes := make([]byte, 0, len(tx.Target.String())+8+8)
	bytes = appendUint64(bytes, tx.Sequence)
	bytes = appendUint64(bytes, uint64(tx.Period))
	bytes = append(bytes, []byte(tx.Target.String())...)
	return bytes
}

// NewChangeSettlementPeriod creates a new signed settlement period change
func NewChangeSettlementPeriod(
	target address.Address,
	newPeriod math.Duration,
	sequence uint64,
	keys []signature.PrivateKey,
) (ChangeSettlementPeriod, error) {
	tx := ChangeSettlementPeriod{
		Target:   target,
		Period:   newPeriod,
		Sequence: sequence,
	}
	sb := tx.SignableBytes()
	for _, key := range keys {
		tx.Signatures = append(tx.Signatures, key.Sign(sb))
	}
	return tx, nil
}

// Validate implements metatx.Transactable
func (tx *ChangeSettlementPeriod) Validate(appI interface{}) (err error) {
	app := appI.(*App)

	if tx.Period < 0 {
		return errors.New("Negative settlement period")
	}
	_, _, _, err = app.getTxAccount(
		tx,
		tx.Target,
		tx.Sequence,
		tx.Signatures,
	)
	if err != nil {
		return err
	}

	return nil
}

// Apply implements metatx.Transactable
func (tx *ChangeSettlementPeriod) Apply(appI interface{}) error {
	app := appI.(*App)
	return app.UpdateState(func(stateI metast.State) (metast.State, error) {
		state := stateI.(*backing.State)
		acct, _ := state.GetAccount(tx.Target, app.blockTime)
		acct.UpdateSettlements(app.blockTime)
		acct.Sequence = tx.Sequence

		fee, err := app.calculateTxFee(tx)
		if err != nil {
			return state, err
		}
		acct.Balance -= fee

		ca := app.blockTime.Add(acct.SettlementSettings.Period)
		acct.SettlementSettings.ChangesAt = &ca
		acct.SettlementSettings.Next = &tx.Period

		state.Accounts[tx.Target.String()] = acct
		return state, nil
	})
}

// GetSource implements sourcer
func (tx *ChangeSettlementPeriod) GetSource(*App) (address.Address, error) {
	return tx.Target, nil
}

// GetSequence implements sequencer
func (tx *ChangeSettlementPeriod) GetSequence() uint64 {
	return tx.Sequence
}
