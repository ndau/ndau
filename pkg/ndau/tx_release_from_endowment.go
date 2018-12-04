package ndau

import (
	metast "github.com/oneiro-ndev/metanode/pkg/meta/state"
	"github.com/oneiro-ndev/ndau/pkg/ndau/backing"
	sv "github.com/oneiro-ndev/system_vars/pkg/system_vars"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	"github.com/oneiro-ndev/ndaumath/pkg/signature"
	math "github.com/oneiro-ndev/ndaumath/pkg/types"
	"github.com/pkg/errors"
)

// GetAccountAddresses returns the account addresses associated with this transaction type.
func (tx *ReleaseFromEndowment) GetAccountAddresses() []string {
	return []string{tx.Destination.String()}
}

// SignableBytes implements Transactable
func (tx *ReleaseFromEndowment) SignableBytes() []byte {
	bytes := make([]byte, 0, tx.Destination.Msgsize()+tx.Qty.Msgsize()+8+8)
	bytes = appendUint64(bytes, tx.Sequence)
	bytes = appendUint64(bytes, uint64(tx.Qty))
	bytes = append(bytes, []byte(tx.Destination.String())...)
	return bytes
}

// NewReleaseFromEndowment constructs a ReleaseFromEndowment transactable.
//
// The caller must ensure that `private` corresponds to a public key listed
// in the `ReleaseFromEndowmentKeys` system variable.
func NewReleaseFromEndowment(
	qty math.Ndau,
	destination address.Address,
	sequence uint64,
	keys []signature.PrivateKey,
) (tx ReleaseFromEndowment) {
	tx = ReleaseFromEndowment{
		Qty:         qty,
		Destination: destination,
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

	_, _, _, err := app.getTxAccount(tx)

	return err
}

// Apply implements metatx.Transactable
func (tx *ReleaseFromEndowment) Apply(appI interface{}) error {
	app := appI.(*App)
	err := app.applyTxDetails(tx)
	if err != nil {
		return err
	}

	return app.UpdateState(func(stateI metast.State) (metast.State, error) {
		var err error
		state := stateI.(*backing.State)

		acct, _ := state.GetAccount(tx.Destination, app.blockTime)
		acct.Balance, err = acct.Balance.Add(tx.Qty)
		if err == nil {
			state.Accounts[tx.Destination.String()] = acct
		}
		return state, err
	})
}

// GetSource implements sourcer
func (tx *ReleaseFromEndowment) GetSource(app *App) (addr address.Address, err error) {
	err = app.System(sv.ReleaseFromEndowmentAddressName, &addr)
	return
}

// GetSequence implements sequencer
func (tx *ReleaseFromEndowment) GetSequence() uint64 {
	return tx.Sequence
}

// GetSignatures implements signeder
func (tx *ReleaseFromEndowment) GetSignatures() []signature.Signature {
	return tx.Signatures
}

// ExtendSignatures implements Signable
func (tx *ReleaseFromEndowment) ExtendSignatures(sa []signature.Signature) {
	tx.Signatures = append(tx.Signatures, sa...)
}
