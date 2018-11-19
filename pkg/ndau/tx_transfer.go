package ndau

import (
	"encoding/binary"

	metast "github.com/oneiro-ndev/metanode/pkg/meta/state"
	"github.com/oneiro-ndev/ndau/pkg/ndau/backing"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	"github.com/oneiro-ndev/ndaumath/pkg/signature"
	math "github.com/oneiro-ndev/ndaumath/pkg/types"
	"github.com/pkg/errors"
)

// NewTransfer creates a new signed transfer transactable
func NewTransfer(
	s address.Address, d address.Address,
	q math.Ndau,
	seq uint64,
	keys []signature.PrivateKey,
) (*Transfer, error) {
	if s == d {
		return nil, errors.New("source may not equal destination")
	}
	tx := &Transfer{
		Source:      s,
		Destination: d,
		Qty:         q,
		Sequence:    seq,
	}
	bytes := tx.SignableBytes()
	for _, key := range keys {
		tx.Signatures = append(tx.Signatures, key.Sign(bytes))
	}

	return tx, nil
}

func appendUint64(b []byte, i uint64) []byte {
	ib := make([]byte, 8)
	binary.BigEndian.PutUint64(ib, i)
	return append(b, ib...)
}

// SignableBytes implements Transactable
func (tx *Transfer) SignableBytes() []byte {
	bytes := make([]byte, 0, tx.Msgsize())
	bytes = append(bytes, tx.Source.String()...)
	bytes = append(bytes, tx.Destination.String()...)
	bytes = appendUint64(bytes, uint64(tx.Qty))
	bytes = appendUint64(bytes, tx.Sequence)
	return bytes
}

// Validate satisfies metatx.Transactable
func (tx *Transfer) Validate(appInt interface{}) error {
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

	dest, _ := state.GetAccount(tx.Destination, app.blockTime)

	if dest.IsNotified(app.blockTime) {
		return errors.New("transfers into notified addresses are invalid")
	}

	return nil
}

// Apply satisfies metatx.Transactable
func (tx *Transfer) Apply(appInt interface{}) error {
	app := appInt.(*App)
	err := app.applyTxDetails(tx)
	if err != nil {
		return err
	}

	state := app.GetState().(*backing.State)

	source, _ := state.GetAccount(tx.Source, app.blockTime)
	dest, _ := state.GetAccount(tx.Destination, app.blockTime)

	err = (&dest.WeightedAverageAge).UpdateWeightedAverageAge(
		app.blockTime.Since(dest.LastWAAUpdate),
		tx.Qty,
		dest.Balance,
	)
	if err != nil {
		return errors.Wrap(err, "update waa")
	}
	dest.LastWAAUpdate = app.blockTime

	dest.Balance += tx.Qty
	if source.SettlementSettings.Period != 0 {
		dest.Settlements = append(dest.Settlements, backing.Settlement{
			Qty:    tx.Qty,
			Expiry: app.blockTime.Add(source.SettlementSettings.Period),
		})
	}

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
func (tx *Transfer) GetSource(*App) (address.Address, error) {
	return tx.Source, nil
}

// Withdrawal implements withdrawer
func (tx *Transfer) Withdrawal() math.Ndau {
	return tx.Qty
}

// GetSequence implements sequencer
func (tx *Transfer) GetSequence() uint64 {
	return tx.Sequence
}

// GetSignatures implements signeder
func (tx *Transfer) GetSignatures() []signature.Signature {
	return tx.Signatures
}

// ExtendSignatures implements Signable
func (tx *Transfer) ExtendSignatures(sa []signature.Signature) {
	tx.Signatures = append(tx.Signatures, sa...)
}
