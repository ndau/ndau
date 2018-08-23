package ndau

import (
	"fmt"

	metast "github.com/oneiro-ndev/metanode/pkg/meta/state"
	"github.com/oneiro-ndev/ndau/pkg/ndau/backing"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	"github.com/oneiro-ndev/signature/pkg/signature"

	"github.com/pkg/errors"
)

// SignableBytes implements Transactable
func (tx *ChangeValidation) SignableBytes() []byte {
	blen := 0
	blen += address.AddrLength
	for _, key := range tx.NewKeys {
		blen += key.Size()
	}
	blen += len(tx.ValidationScript)
	blen += 8 // sequence
	bytes := make([]byte, 0, blen)

	bytes = appendUint64(bytes, tx.Sequence)
	bytes = append(bytes, tx.Target.String()...)
	for _, key := range tx.NewKeys {
		bytes = append(bytes, key.Bytes()...)
	}
	bytes = append(bytes, tx.ValidationScript...)

	return bytes
}

// NewChangeValidation creates a new signed transfer key from its data and a private key
func NewChangeValidation(
	target address.Address,
	newKeys []signature.PublicKey,
	validationScript []byte,
	sequence uint64,
	privates []signature.PrivateKey,
) ChangeValidation {
	tx := ChangeValidation{
		Target:           target,
		NewKeys:          newKeys,
		ValidationScript: validationScript,
		Sequence:         sequence,
	}
	for _, private := range privates {
		tx.Signatures = append(tx.Signatures, private.Sign(tx.SignableBytes()))
	}
	return tx
}

// Validate implements metatx.Transactable
func (tx *ChangeValidation) Validate(appI interface{}) (err error) {
	tx.Target, err = address.Validate(tx.Target.String())
	if err != nil {
		return
	}

	// business rule: there must be at least 1 and no more than a const
	// transfer keys set in this tx
	if len(tx.NewKeys) < 1 || len(tx.NewKeys) > backing.MaxKeysInAccount {
		return fmt.Errorf("Expect between 1 and %d transfer keys; got %d", backing.MaxKeysInAccount, len(tx.NewKeys))
	}

	if len(tx.ValidationScript) > 0 && !IsChaincode(tx.ValidationScript) {
		return errors.New("Validation script must be chaincode")
	}

	app := appI.(*App)
	_, _, _, err = app.getTxAccount(
		tx,
		tx.Target,
		tx.Sequence,
		tx.Signatures,
	)
	if err != nil {
		return err
	}

	// get the target address kind for later use:
	// we need to generate addresses for the signing key, to verify it matches
	// the actual ownership key, if used, and for the new transfer key,
	// to ensure it's not equal to the actual ownership key
	kind := address.Kind(string(tx.Target.String()[2]))
	if !address.IsValidKind(kind) {
		return fmt.Errorf("Target has invalid address kind: %s", kind)
	}

	// per-key validation
	for _, tk := range tx.NewKeys {
		// new transfer key must not equal ownership key
		ntAddr, err := address.Generate(kind, tk.Bytes())
		if err != nil {
			return errors.Wrap(err, "Failed to generate address from new transfer key")
		}
		if tx.Target.String() == ntAddr.String() {
			return fmt.Errorf("New transfer key must not equal ownership key")
		}
	}

	return
}

// Apply implements metatx.Transactable
func (tx *ChangeValidation) Apply(appI interface{}) error {
	app := appI.(*App)
	return app.UpdateState(func(stateI metast.State) (metast.State, error) {
		state := stateI.(*backing.State)

		ad, hasAd := state.GetAccount(tx.Target, app.blockTime)
		if !hasAd {
			ad = backing.NewAccountData(app.blockTime)
		}
		ad.Sequence = tx.Sequence

		ad.TransferKeys = tx.NewKeys
		ad.ValidationScript = tx.ValidationScript

		state.Accounts[tx.Target.String()] = ad
		return state, nil
	})
}
