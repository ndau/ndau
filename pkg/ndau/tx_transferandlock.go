package ndau

import (
	metast "github.com/oneiro-ndev/metanode/pkg/meta/state"
	"github.com/oneiro-ndev/ndau/pkg/ndau/backing"
	sv "github.com/oneiro-ndev/ndau/pkg/ndau/system_vars"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	"github.com/oneiro-ndev/ndaumath/pkg/eai"
	"github.com/oneiro-ndev/ndaumath/pkg/signature"
	math "github.com/oneiro-ndev/ndaumath/pkg/types"
	"github.com/pkg/errors"
)

// NewTransferAndLock creates a new signed transferAndLock transactable
func NewTransferAndLock(
	s address.Address,
	d address.Address,
	q math.Ndau,
	p math.Duration,
	seq uint64,
	keys []signature.PrivateKey,
) (*TransferAndLock, error) {
	if s == d {
		return nil, errors.New("source may not equal destination")
	}
	tx := &TransferAndLock{
		Source:      s,
		Destination: d,
		Qty:         q,
		Period:      p,
		Sequence:    seq,
	}
	bytes := tx.SignableBytes()
	for _, key := range keys {
		tx.Signatures = append(tx.Signatures, key.Sign(bytes))
	}

	return tx, nil
}

// SignableBytes implements Transactable
func (tx *TransferAndLock) SignableBytes() []byte {
	bytes := make([]byte, 0, tx.Msgsize())
	bytes = append(bytes, tx.Source.String()...)
	bytes = append(bytes, tx.Destination.String()...)
	bytes = appendUint64(bytes, uint64(tx.Qty))
	bytes = appendUint64(bytes, uint64(tx.Period))
	bytes = appendUint64(bytes, tx.Sequence)
	return bytes
}

// Validate satisfies metatx.Transactable
func (tx *TransferAndLock) Validate(appInt interface{}) error {
	app := appInt.(*App)
	state := app.GetState().(*backing.State)

	if tx.Qty <= math.Ndau(0) {
		return errors.New("invalid transfer: Qty not positive")
	}

	if tx.Source == tx.Destination {
		return errors.New("invalid transfer: source == destination")
	}

	source, _, _, err := app.getTxAccount(tx)
	if err != nil {
		return err
	}

	if source.IsLocked(app.blockTime) {
		return errors.New("source is locked")
	}

	_, exists := state.GetAccount(tx.Destination, app.blockTime)

	// TransferAndLock cannot be sent to an existing account
	if exists {
		return errors.New("invalid TransferAndLock: cannot be sent to an existing account")
	}

	return nil
}

// Apply satisfies metatx.Transactable
func (tx *TransferAndLock) Apply(appInt interface{}) error {
	app := appInt.(*App)
	err := app.applyTxDetails(tx)
	if err != nil {
		return err
	}

	lockedBonusRateTable := eai.RateTable{}
	err = app.System(sv.LockedRateTableName, &lockedBonusRateTable)
	if err != nil {
		return err
	}

	state := app.GetState().(*backing.State)

	source, _ := state.GetAccount(tx.Source, app.blockTime)
	// we know dest is a new account so WAA, WAA update and EAI update times are set properly
	dest, _ := state.GetAccount(tx.Destination, app.blockTime)
	dest.Balance = tx.Qty
	if source.SettlementSettings.Period != 0 {
		dest.Settlements = append(dest.Settlements, backing.Settlement{
			Qty:    tx.Qty,
			Expiry: app.blockTime.Add(source.SettlementSettings.Period),
		})
	}
	dest.Lock = backing.NewLock(tx.Period, lockedBonusRateTable)

	return app.UpdateState(func(stateI metast.State) (metast.State, error) {
		state := stateI.(*backing.State)

		state.Accounts[tx.Destination.String()] = dest
		if source.Balance > 0 {
			state.Accounts[tx.Source.String()] = source
		} else {
			delete(state.Accounts, tx.Source.String())
		}

		return state, nil
	})
}

// GetSource implements sourcer
func (tx *TransferAndLock) GetSource(*App) (address.Address, error) {
	return tx.Source, nil
}

// Withdrawal implements withdrawer
func (tx *TransferAndLock) Withdrawal() math.Ndau {
	return tx.Qty
}

// GetSequence implements sequencer
func (tx *TransferAndLock) GetSequence() uint64 {
	return tx.Sequence
}

// GetSignatures implements signeder
func (tx *TransferAndLock) GetSignatures() []signature.Signature {
	return tx.Signatures
}

// ExtendSignatures implements Signable
func (tx *TransferAndLock) ExtendSignatures(sa []signature.Signature) {
	tx.Signatures = append(tx.Signatures, sa...)
}
