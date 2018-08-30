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
	t := &Transfer{
		Source:      s,
		Destination: d,
		Qty:         q,
		Sequence:    seq,
	}
	bytes := t.SignableBytes()
	for _, key := range keys {
		t.Signatures = append(t.Signatures, key.Sign(bytes))
	}

	return t, nil
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

func (tx *Transfer) calculateTxFee() math.Ndau {
	// TODO: perform a real calculation here
	return math.Ndau(0)
}

func (tx *Transfer) calculateSIB() math.Ndau {
	// TODO: perform a real lookup here
	return math.Ndau(0)
}

func (tx *Transfer) calculateQtyFromSource() (math.Ndau, error) {
	var err error
	fromSource := tx.Qty
	fromSource, err = fromSource.Add(tx.calculateTxFee())
	if err != nil {
		return math.Ndau(0), errors.Wrap(err, "calculating total from source")
	}
	fromSource, err = fromSource.Add(tx.calculateSIB())
	if err != nil {
		return math.Ndau(0), errors.Wrap(err, "calculating total from source")
	}

	return fromSource, nil
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

	source, _, _, err := app.getTxAccount(
		tx,
		tx.Source,
		tx.Sequence,
		tx.Signatures,
	)
	if err != nil {
		return err
	}

	if source.IsLocked(app.blockTime) {
		return errors.New("source is locked")
	}

	// the source update doesn't get persisted this time because this method is read-only
	source.UpdateSettlements(app.blockTime)
	availableBalance, err := source.AvailableBalance()
	if err != nil {
		return err
	}

	fromSource, err := tx.calculateQtyFromSource()
	if err != nil {
		return err
	}
	if availableBalance.Compare(fromSource) < 0 {
		return errors.New("insufficient balance in source")
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
	state := app.GetState().(*backing.State)

	source, _ := state.GetAccount(tx.Source, app.blockTime)
	dest, _ := state.GetAccount(tx.Destination, app.blockTime)

	// this source update will get persisted if the method exits without error
	source.UpdateSettlements(app.blockTime)

	err := (&dest.WeightedAverageAge).UpdateWeightedAverageAge(
		app.blockTime.Since(dest.LastWAAUpdate),
		tx.Qty,
		dest.Balance,
	)
	if err != nil {
		return errors.Wrap(err, "update waa")
	}
	dest.LastWAAUpdate = app.blockTime

	fromSource, err := tx.calculateQtyFromSource()
	if err != nil {
		return errors.Wrap(err, "calc qty to take from source")
	}
	source.Balance -= fromSource
	source.Sequence = tx.Sequence

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

// CalculateTxFee implements metatx.Transactable
func (tx *Transfer) CalculateTxFee(appI interface{}) (math.Ndau, error) {
	app := appI.(*App)
	return app.calculateTxFee(tx)
}
