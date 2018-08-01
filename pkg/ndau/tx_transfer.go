package ndau

import (
	"encoding/binary"
	"time"

	"github.com/pkg/errors"

	metast "github.com/oneiro-ndev/metanode/pkg/meta/state"
	"github.com/oneiro-ndev/ndau/pkg/ndau/backing"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	math "github.com/oneiro-ndev/ndaumath/pkg/types"
	"github.com/oneiro-ndev/signature/pkg/signature"
)

// NewTransfer creates a new signed transfer transactable
func NewTransfer(
	s address.Address, d address.Address,
	q math.Ndau,
	seq uint64,
	key signature.PrivateKey,
) (*Transfer, error) {
	if s == d {
		return nil, errors.New("source may not equal destination")
	}
	ts, err := math.TimestampFrom(time.Now())
	if err != nil {
		return nil, err
	}
	t := &Transfer{
		Timestamp:   ts,
		Source:      s,
		Destination: d,
		Qty:         q,
		Sequence:    seq,
	}
	bytes := t.SignableBytes()
	t.Signature, err = key.Sign(bytes).Marshal()

	return t, err
}

func appendUint64(b []byte, i uint64) []byte {
	ib := make([]byte, 8)
	binary.BigEndian.PutUint64(b, i)
	return append(b, ib...)
}

// SignableBytes implements Transactable
func (t *Transfer) SignableBytes() []byte {
	bytes := make([]byte, 8+8, t.Msgsize()+8+8)
	binary.BigEndian.PutUint64(bytes[0:8], uint64(t.Timestamp))
	binary.BigEndian.PutUint64(bytes[8:16], uint64(t.Qty))
	bytes = append(bytes, t.Source.String()...)
	bytes = append(bytes, t.Destination.String()...)
	bytes = appendUint64(bytes, t.Sequence)
	return bytes
}

func (t *Transfer) signature(private signature.PrivateKey) ([]byte, error) {
	bytes := t.SignableBytes()

	sigB, err := private.Sign(bytes).Marshal()
	if err != nil {
		return nil, err
	}
	return sigB, nil
}

func (t *Transfer) calculateTxFee() math.Ndau {
	// TODO: perform a real calculation here
	return math.Ndau(0)
}

func (t *Transfer) calculateSIB() math.Ndau {
	// TODO: perform a real lookup here
	return math.Ndau(0)
}

func (t *Transfer) calculateQtyFromSource() (math.Ndau, error) {
	var err error
	fromSource := t.Qty
	fromSource, err = fromSource.Add(t.calculateTxFee())
	if err != nil {
		return math.Ndau(0), errors.Wrap(err, "calculating total from source")
	}
	fromSource, err = fromSource.Add(t.calculateSIB())
	if err != nil {
		return math.Ndau(0), errors.Wrap(err, "calculating total from source")
	}

	return fromSource, nil
}

// Validate satisfies metatx.Transactable
func (t *Transfer) Validate(appInt interface{}) error {
	app := appInt.(*App)
	state := app.GetState().(*backing.State)

	if t.Qty <= math.Ndau(0) {
		return errors.New("invalid transfer: Qty not positive")
	}

	if t.Source == t.Destination {
		return errors.New("invalid transfer: source == destination")
	}

	source := state.Accounts[t.Source.String()]
	if source.IsLocked(app.blockTime) {
		return errors.New("source is locked")
	}

	if source.TransferKey == nil {
		return errors.New("source.TransferKey not set")
	}
	publicKey := *source.TransferKey

	tBytes := t.SignableBytes()
	sig := signature.Signature{}
	err := (&sig).Unmarshal(t.Signature)
	if err != nil {
		return errors.Wrap(err, "unmarshal signature")
	}
	if !publicKey.Verify(tBytes, sig) {
		return errors.New("invalid signature")
	}

	if t.Sequence <= source.Sequence {
		return errors.New("sequence number too low")
	}

	// the source update doesn't get persisted this time because this method is read-only
	source.UpdateSettlement(app.blockTime)

	fromSource, err := t.calculateQtyFromSource()
	if err != nil {
		return err
	}
	if source.Balance.Compare(fromSource) < 0 {
		return errors.New("insufficient balance in source")
	}

	dest := state.Accounts[t.Destination.String()]

	if dest.IsNotified(app.blockTime) {
		return errors.New("transfers into notified addresses are invalid")
	}

	return nil
}

// Apply satisfies metatx.Transactable
func (t *Transfer) Apply(appInt interface{}) error {
	app := appInt.(*App)
	state := app.GetState().(*backing.State)

	source := state.Accounts[t.Source.String()]
	dest, hasDest := state.Accounts[t.Destination.String()]
	if !hasDest {
		dest = backing.NewAccountData(app.blockTime)
	}

	// this source update will get persisted if the method exits without error
	source.UpdateSettlement(app.blockTime)

	err := (&dest.WeightedAverageAge).UpdateWeightedAverageAge(
		app.blockTime.Since(dest.LastWAAUpdate),
		t.Qty,
		dest.Balance,
	)
	if err != nil {
		return errors.Wrap(err, "update waa")
	}
	dest.LastWAAUpdate = app.blockTime

	fromSource, err := t.calculateQtyFromSource()
	if err != nil {
		return errors.Wrap(err, "calc qty to take from source")
	}
	source.Balance -= fromSource
	source.Sequence = t.Sequence
	if source.SettlementSettings.Period == 0 {
		dest.Balance += t.Qty
	} else {
		dest.Settlements = append(dest.Settlements, backing.Settlement{
			Qty:    t.Qty,
			Expiry: app.blockTime.Add(source.SettlementSettings.Period),
		})
	}

	return app.UpdateState(func(stateI metast.State) (metast.State, error) {
		state := stateI.(*backing.State)

		state.Accounts[t.Destination.String()] = dest
		if source.Balance > 0 {
			state.Accounts[t.Source.String()] = source
		} else {
			delete(state.Accounts, t.Source.String())
		}

		return state, nil
	})
}
