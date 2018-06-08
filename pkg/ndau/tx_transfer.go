package ndau

import (
	"encoding/binary"

	math "github.com/oneiro-ndev/ndaumath/pkg/types"
	"github.com/oneiro-ndev/ndaunode/pkg/ndau/backing"
	"github.com/oneiro-ndev/signature/pkg/signature"
	"github.com/pkg/errors"
)

// NewTransfer creates a new signed transfer transactable
func NewTransfer(ts math.Timestamp, s string, d string, q math.Ndau, seq uint64, al signature.Algorithm, privateKey signature.Key) (*Transfer, error) {
	if s == d {
		return nil, errors.New("source may not equal destination")
	}
	if al == nil {
		return nil, errors.New("Nil Algorithm")
	}
	t := &Transfer{
		Timestamp:   ts,
		Source:      s,
		Destination: d,
		Qty:         q,
		Sequence:    seq,
	}
	sig, err := t.signature(al, privateKey)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to sign NewTransfer")
	}
	t.Signature = sig
	return t, nil
}

func appendUint64(b []byte, i uint64) []byte {
	ib := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, i)
	return append(b, ib...)
}

func (t *Transfer) signableBytes() ([]byte, error) {
	var err error
	bytes := make([]byte, t.Msgsize())
	bytes, err = t.Timestamp.MarshalMsg(bytes)
	if err != nil {
		return nil, err
	}
	bytes = append(bytes, t.Source...)
	bytes = append(bytes, t.Destination...)
	bytes, err = t.Qty.MarshalMsg(bytes)
	if err != nil {
		return nil, err
	}
	bytes = appendUint64(bytes, t.Sequence)
	return bytes, nil
}

func (t *Transfer) signature(al signature.Algorithm, private signature.Key) ([]byte, error) {
	bytes, err := t.signableBytes()
	if err != nil {
		return nil, err
	}
	sig := al.Sign(private, bytes)
	sigB, err := signature.MarshalSignature(al, sig)
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

// IsValid satisfies metatx.Transactable
func (t *Transfer) IsValid(appInt interface{}) error {
	app := appInt.(*App)
	state := app.GetState().(*backing.State)

	if t.Source == t.Destination {
		return errors.New("invalid transfer: source == destination")
	}

	source, err := state.GetAccount(app.GetDB(), t.Source)
	if err != nil {
		return errors.Wrap(err, "Source")
	}
	if source.IsLocked(app.blockTime) {
		return errors.New("source is locked")
	}

	if source.TransferKey == nil {
		return errors.New("source.TransferKey not set")
	}
	algorithm, publicKey, err := signature.UnmarshalKey(source.TransferKey)
	if err != nil {
		return errors.Wrap(err, "source.TransferKey")
	}
	tBytes, err := t.signableBytes()
	if err != nil {
		return errors.Wrap(err, "signable bytes")
	}
	if !algorithm.Verify(publicKey, tBytes, t.Signature) {
		return errors.New("invalid signature")
	}

	if t.Sequence <= source.Sequence {
		return errors.New("sequence number too low")
	}

	fromSource, err := t.calculateQtyFromSource()
	if err != nil {
		return err
	}
	if source.Balance.Compare(fromSource) < 0 {
		return errors.New("insufficient balance in source")
	}

	dest, err := state.GetAccount(app.GetDB(), t.Destination)
	if err != nil {
		return errors.Wrap(err, "Destination")
	}

	if dest.IsNotified(app.blockTime) {
		return errors.New("transfers into notified addresses are invalid")
	}

	return nil
}

// Apply satisfies metatx.Transactable
func (t *Transfer) Apply(appInt interface{}) error {
	app := appInt.(*App)
	state := app.GetState().(*backing.State)

	source, err := state.GetAccount(app.GetDB(), t.Source)
	if err != nil {
		return errors.Wrap(err, "Source")
	}
	dest, err := state.GetAccount(app.GetDB(), t.Destination)
	if err != nil {
		return errors.Wrap(err, "Destination")
	}

	fromSource, err := t.calculateQtyFromSource()
	if err != nil {
		return errors.Wrap(err, "calc qty to take from source")
	}
	source.Balance -= fromSource
	source.Sequence = t.Sequence
	dest.Balance += t.Qty

	// fail safe: update the dest before the source, so if only one
	// fails, we're crediting too much money instead of burning it
	err = state.UpdateAccount(app.GetDB(), t.Destination, dest)
	if err != nil {
		return errors.Wrap(err, "update Destination")
	}
	err = state.UpdateAccount(app.GetDB(), t.Source, source)
	if err != nil {
		return errors.Wrap(err, "update Source")
	}

	return nil
}
