package ndau

import (
	"encoding/binary"
	"fmt"

	metast "github.com/oneiro-ndev/metanode/pkg/meta/state"
	"github.com/oneiro-ndev/ndau/pkg/ndau/backing"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	"github.com/oneiro-ndev/signature/pkg/signature"

	"github.com/pkg/errors"
)

// SignableBytes implements Transactable
func (ct *ChangeValidation) SignableBytes() []byte {
	blen := 0
	blen += address.AddrLength
	for _, key := range ct.NewKeys {
		blen += key.Size()
	}
	blen += 8
	bytes := make([]byte, 8, blen)

	binary.BigEndian.PutUint64(bytes, ct.Sequence)
	bytes = append(bytes, ct.Target.String()...)
	for _, key := range ct.NewKeys {
		bytes = append(bytes, key.Bytes()...)
	}

	return bytes
}

// NewChangeValidation creates a new signed transfer key from its data and a private key
func NewChangeValidation(
	target address.Address,
	newKeys []signature.PublicKey,
	sequence uint64,
	privates []signature.PrivateKey,
) ChangeValidation {
	ct := ChangeValidation{
		Target:   target,
		NewKeys:  newKeys,
		Sequence: sequence,
	}
	for _, private := range privates {
		ct.Signatures = append(ct.Signatures, private.Sign(ct.SignableBytes()))
	}
	return ct
}

// Validate implements metatx.Transactable
func (ct *ChangeValidation) Validate(appI interface{}) (err error) {
	ct.Target, err = address.Validate(ct.Target.String())
	if err != nil {
		return
	}

	// business rule: there must be at least 1 and no more than a const
	// transfer keys set in this tx
	if len(ct.NewKeys) < 1 || len(ct.NewKeys) > backing.MaxKeysInAccount {
		return fmt.Errorf("Expect between 1 and %d transfer keys; got %d", backing.MaxKeysInAccount, len(ct.NewKeys))
	}

	app := appI.(*App)
	_, _, err = app.GetState().(*backing.State).GetValidAccount(
		ct.Target,
		app.blockTime,
		ct.Sequence,
		ct.SignableBytes(),
		ct.Signatures,
	)
	if err != nil {
		return err
	}

	// get the target address kind for later use:
	// we need to generate addresses for the signing key, to verify it matches
	// the actual ownership key, if used, and for the new transfer key,
	// to ensure it's not equal to the actual ownership key
	kind := address.Kind(string(ct.Target.String()[2]))
	if !address.IsValidKind(kind) {
		return fmt.Errorf("Target has invalid address kind: %s", kind)
	}

	// per-key validation
	for _, tk := range ct.NewKeys {
		// new transfer key must not equal ownership key
		ntAddr, err := address.Generate(kind, tk.Bytes())
		if err != nil {
			return errors.Wrap(err, "Failed to generate address from new transfer key")
		}
		if ct.Target.String() == ntAddr.String() {
			return fmt.Errorf("New transfer key must not equal ownership key")
		}
	}

	return
}

// Apply implements metatx.Transactable
func (ct *ChangeValidation) Apply(appI interface{}) error {
	app := appI.(*App)
	return app.UpdateState(func(stateI metast.State) (metast.State, error) {
		state := stateI.(*backing.State)

		ad, hasAd := state.GetAccount(ct.Target, app.blockTime)
		if !hasAd {
			ad = backing.NewAccountData(app.blockTime)
		}
		ad.Sequence = ct.Sequence

		ad.TransferKeys = ct.NewKeys

		state.Accounts[ct.Target.String()] = ad
		return state, nil
	})
}
