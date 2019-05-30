package ndau

import (
	"fmt"

	metast "github.com/oneiro-ndev/metanode/pkg/meta/state"
	"github.com/oneiro-ndev/ndau/pkg/ndau/backing"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	"github.com/oneiro-ndev/ndaumath/pkg/signature"

	"github.com/pkg/errors"
)

// GetAccountAddresses returns the account addresses associated with this transaction type.
func (tx *ChangeValidation) GetAccountAddresses() []string {
	return []string{tx.Target.String()}
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
	_, _, _, err = app.getTxAccount(tx)
	if err != nil {
		return err
	}

	// get the target address kind for later use:
	// we need to generate addresses for the signing key, to verify it matches
	// the actual ownership key, if used, and for the new transfer key,
	// to ensure it's not equal to the actual ownership key
	kind := tx.Target.Kind()
	if !address.IsValidKind(kind) {
		return fmt.Errorf("Target has invalid address kind: %x", kind)
	}

	// per-key validation
	for _, tk := range tx.NewKeys {
		// new transfer key must not equal ownership key
		ntAddr, err := address.Generate(kind, tk.KeyBytes())
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
	err := app.applyTxDetails(tx)
	if err != nil {
		return err
	}

	return app.UpdateState(func(stateI metast.State) (metast.State, error) {
		state := stateI.(*backing.State)
		ad, _ := app.getAccount(tx.Target)

		ad.ValidationKeys = tx.NewKeys
		ad.ValidationScript = tx.ValidationScript

		state.Accounts[tx.Target.String()] = ad
		return state, nil
	})
}

// GetSource implements Sourcer
func (tx *ChangeValidation) GetSource(*App) (address.Address, error) {
	return tx.Target, nil
}

// GetSequence implements Sequencer
func (tx *ChangeValidation) GetSequence() uint64 {
	return tx.Sequence
}

// GetSignatures implements Signeder
func (tx *ChangeValidation) GetSignatures() []signature.Signature {
	return tx.Signatures
}

// ExtendSignatures implements Signable
func (tx *ChangeValidation) ExtendSignatures(sa []signature.Signature) {
	tx.Signatures = append(tx.Signatures, sa...)
}
