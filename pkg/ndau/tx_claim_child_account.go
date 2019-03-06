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
func (tx *ClaimChildAccount) GetAccountAddresses() []string {
	return []string{tx.Target.String(), tx.Child.String()}
}

// Validate returns nil if tx is valid, or an error
func (tx *ClaimChildAccount) Validate(appI interface{}) error {
	// Ensure the target and child address are valid.
	_, err := address.Validate(tx.Target.String())
	if err != nil {
		return errors.Wrap(err,
			fmt.Sprintf("Target account address invalid: %s", tx.Target.String()))
	}
	_, err = address.Validate(tx.Child.String())
	if err != nil {
		return errors.Wrap(err,
			fmt.Sprintf("Child account address invalid: %s", tx.Child.String()))
	}

	// Verify that the child ownership key submitted generates the child address being claimed.
	childKind := tx.Child.Kind()
	if !address.IsValidKind(childKind) {
		return fmt.Errorf("Child account %s has invalid address kind: %x",
			tx.Child.String(), childKind)
	}
	childOwnershipAddress, err := address.Generate(childKind, tx.ChildOwnership.KeyBytes())
	if err != nil {
		return errors.Wrap(err, "Unable to generate address for child ownership key")
	}
	if tx.Child.String() != childOwnershipAddress.String() {
		return errors.New("Child ownership key and address do not match")
	}

	// The child ownership signature should properly sign the signable bytes
	// using the child ownership key.
	if !tx.ChildSignature.Verify(tx.SignableBytes(), tx.ChildOwnership) {
		return errors.New("Child signature unable to sign transaction")
	}

	// Similar to ClaimAccount tx: there must be at least 1 and no more than a const transfer
	// keys set in this tx.
	if len(tx.ChildValidationKeys) < 1 || len(tx.ChildValidationKeys) > backing.MaxKeysInAccount {
		return fmt.Errorf("Expect between 1 and %d transfer keys; got %d",
			backing.MaxKeysInAccount, len(tx.ChildValidationKeys))
	}

	// No transfer key may be equal to the ownership key.
	for _, tk := range tx.ChildValidationKeys {
		tkAddress, err := address.Generate(childKind, tk.KeyBytes())
		if err != nil {
			return errors.Wrap(err, "Unable to generate address for transfer key")
		}
		if tkAddress.String() == childOwnershipAddress.String() {
			return errors.New("Child ownership key may not be used as a transfer key")
		}
	}

	// Ensure the validation scripts are chaincode.
	if len(tx.ChildValidationScript) > 0 && !IsChaincode(tx.ChildValidationScript) {
		return errors.New("Child validation script must be chaincode")
	}

	app := appI.(*App)
	child, _, _, err := app.getTxAccount(tx)
	if err != nil {
		return err
	}

	// Ensure the child account is not already claimed.
	// This consequently also ensures that:
	// - We won't have ancestry loops
	// - You cannot claim a locked child account
	if len(child.ValidationKeys) > 0 {
		return errors.New("Cannot claim a child account that is already claimed")
	}

	return nil
}

// Apply applies this tx if no error occurs
func (tx *ClaimChildAccount) Apply(appI interface{}) error {
	app := appI.(*App)
	err := app.applyTxDetails(tx)
	if err != nil {
		return err
	}

	return app.UpdateState(func(stI metast.State) (metast.State, error) {
		st := stI.(*backing.State)

		child, _ := app.getAccount(tx.Child)
		child.ValidationKeys = tx.ChildValidationKeys
		child.ValidationScript = tx.ChildValidationScript

		target, _ := app.getAccount(tx.Target)
		child.Parent = &tx.Target
		if target.Progenitor == nil {
			child.Progenitor = &tx.Target
		} else {
			child.Progenitor = target.Progenitor
		}

		st.Accounts[tx.Child.String()] = child

		return st, nil
	})
}

// GetSource implements sourcer
func (tx *ClaimChildAccount) GetSource(*App) (address.Address, error) {
	return tx.Target, nil
}

// GetSequence implements sequencer
func (tx *ClaimChildAccount) GetSequence() uint64 {
	return tx.Sequence
}

// GetSignatures implements signeder
func (tx *ClaimChildAccount) GetSignatures() []signature.Signature {
	return tx.Signatures
}
