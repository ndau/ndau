package ndau

import (
	"encoding/binary"
	"fmt"

	metast "github.com/oneiro-ndev/metanode/pkg/meta/state"
	metatx "github.com/oneiro-ndev/metanode/pkg/meta/transaction"
	"github.com/oneiro-ndev/ndau/pkg/ndau/backing"
	sv "github.com/oneiro-ndev/ndau/pkg/ndau/system_vars"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	math "github.com/oneiro-ndev/ndaumath/pkg/types"
	"github.com/oneiro-ndev/signature/pkg/signature"
	"github.com/pkg/errors"
)

// SignableBytes implements Transactable
func (rfe *ReleaseFromEndowment) SignableBytes() []byte {
	bytes := make([]byte, 8+8, rfe.Destination.Msgsize()+rfe.Qty.Msgsize()+rfe.TxFeeAcct.Msgsize()+8+8)
	binary.BigEndian.PutUint64(bytes[0:8], rfe.Sequence)
	binary.BigEndian.PutUint64(bytes[8:16], uint64(rfe.Qty))
	bytes = append(bytes, []byte(rfe.Destination.String())...)
	bytes = append(bytes, []byte(rfe.TxFeeAcct.String())...)
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
	private signature.PrivateKey,
) (rfe ReleaseFromEndowment) {
	rfe = ReleaseFromEndowment{
		Qty:         qty,
		Destination: destination,
		TxFeeAcct:   txFeeAcct,
		Sequence:    sequence,
	}
	rfe.Signature = metatx.Sign(&rfe, private)
	return rfe
}

// Validate implements metatx.Transactable
func (rfe *ReleaseFromEndowment) Validate(appI interface{}) error {
	app := appI.(*App)

	if rfe.Qty <= 0 {
		return errors.New("RFE qty may not be <= 0")
	}

	state := app.GetState().(*backing.State)
	txAcct, hasAcct := state.Accounts[rfe.TxFeeAcct.String()]
	if !hasAcct {
		return errors.New("TxFeeAcct does not exist")
	}
	if rfe.Sequence <= txAcct.Sequence {
		return errors.New("Sequence too low")
	}
	sb := rfe.SignableBytes()
	if txAcct.TransferKey == nil {
		return errors.New("TxFeeAcct transfer key not set")
	}
	if !txAcct.TransferKey.Verify(sb, rfe.Signature) {
		return errors.New("TxFeeAcct TransferKey does not validate Signature")
	}

	rfeKeys := make(sv.ReleaseFromEndowmentKeys, 0)
	err := app.System(sv.ReleaseFromEndowmentKeysName, &rfeKeys)
	if err != nil {
		return errors.Wrap(err, "RFE.Validate app.System err")
	}
	valid := false
	for _, public := range rfeKeys {
		if public.Verify(sb, rfe.Signature) {
			valid = true
			break
		}
	}
	if !valid {
		return fmt.Errorf("No public key in %s verifies RFE signature", sv.ReleaseFromEndowmentKeysName)
	}

	return nil
}

// Apply implements metatx.Transactable
func (rfe *ReleaseFromEndowment) Apply(appI interface{}) error {
	app := appI.(*App)
	return app.UpdateState(func(stateI metast.State) (metast.State, error) {
		var err error
		state := stateI.(*backing.State)

		txAcct := state.Accounts[rfe.TxFeeAcct.String()]
		txAcct.Sequence = rfe.Sequence
		state.Accounts[rfe.TxFeeAcct.String()] = txAcct

		acct := state.Accounts[rfe.Destination.String()]
		acct.Balance, err = acct.Balance.Add(rfe.Qty)
		if err == nil {
			state.Accounts[rfe.Destination.String()] = acct
		}
		return state, err
	})
}
