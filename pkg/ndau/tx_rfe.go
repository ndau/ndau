package ndau

import (
	"fmt"

	metast "github.com/oneiro-ndev/metanode/pkg/meta/state"
	"github.com/oneiro-ndev/ndau/pkg/ndau/backing"
	sv "github.com/oneiro-ndev/ndau/pkg/ndau/system_vars"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	math "github.com/oneiro-ndev/ndaumath/pkg/types"
	"github.com/oneiro-ndev/signature/pkg/signature"
	"github.com/pkg/errors"
)

// SignableBytes implements Transactable
func (tx *ReleaseFromEndowment) SignableBytes() []byte {
	bytes := make([]byte, 0, tx.Destination.Msgsize()+tx.Qty.Msgsize()+tx.TxFeeAcct.Msgsize()+8+8)
	bytes = appendUint64(bytes, tx.Sequence)
	bytes = appendUint64(bytes, uint64(tx.Qty))
	bytes = append(bytes, []byte(tx.Destination.String())...)
	bytes = append(bytes, []byte(tx.TxFeeAcct.String())...)
	return bytes
}

// NewReleaseFromEndowment constructs a ReleaseFromEndowment transactable.
//
// The caller must ensure that `private` corresponds to a public key listed
// in the `ReleaseFromEndowmentKeys` system variable.
func NewReleaseFromEndowment(
	qty math.Ndau,
	destination, txFeeAcct address.Address,
	sequence uint64,
	keys []signature.PrivateKey,
) (tx ReleaseFromEndowment) {
	tx = ReleaseFromEndowment{
		Qty:         qty,
		Destination: destination,
		TxFeeAcct:   txFeeAcct,
		Sequence:    sequence,
	}
	for _, key := range keys {
		tx.Signatures = append(tx.Signatures, key.Sign(tx.SignableBytes()))
	}
	return tx
}

// Validate implements metatx.Transactable
func (tx *ReleaseFromEndowment) Validate(appI interface{}) error {
	app := appI.(*App)

	if tx.Qty <= 0 {
		return errors.New("RFE qty may not be <= 0")
	}

	_, hasAcct, _, err := app.getTxAccount(
		tx,
		tx.TxFeeAcct,
		tx.Sequence,
		tx.Signatures,
	)
	if err != nil {
		return err
	}

	if !hasAcct {
		return errors.New("TxFeeAcct does not exist")
	}

	rfeKeys := make(sv.ReleaseFromEndowmentKeys, 0)
	err = app.System(sv.ReleaseFromEndowmentKeysName, &rfeKeys)
	if err != nil {
		return errors.Wrap(err, "RFE.Validate app.System err")
	}
	// all signatures must be validated by keys in the rfeKeys list
	valid := true
	for _, sig := range tx.Signatures {
		match := false
		for _, public := range rfeKeys {
			if public.Verify(tx.SignableBytes(), sig) {
				match = true
				break
			}
		}
		valid = valid && match
	}
	if !valid {
		return fmt.Errorf("No public key in %s verifies RFE signature", sv.ReleaseFromEndowmentKeysName)
	}

	return nil
}

// Apply implements metatx.Transactable
func (tx *ReleaseFromEndowment) Apply(appI interface{}) error {
	app := appI.(*App)
	return app.UpdateState(func(stateI metast.State) (metast.State, error) {
		var err error
		state := stateI.(*backing.State)

		txAcct, _ := state.GetAccount(tx.TxFeeAcct, app.blockTime)
		txAcct.Sequence = tx.Sequence

		fee, err := app.calculateTxFee(tx)
		if err != nil {
			return state, err
		}
		txAcct.Balance -= fee
		state.Accounts[tx.TxFeeAcct.String()] = txAcct

		acct, _ := state.GetAccount(tx.Destination, app.blockTime)
		acct.Balance, err = acct.Balance.Add(tx.Qty)
		if err == nil {
			state.Accounts[tx.Destination.String()] = acct
		}
		return state, err
	})
}

// GetSource implements sourcer
func (tx *ReleaseFromEndowment) GetSource(*App) (address.Address, error) {
	return tx.TxFeeAcct, nil
}

// GetSequence implements sequencer
func (tx *ReleaseFromEndowment) GetSequence() uint64 {
	return tx.Sequence
}
