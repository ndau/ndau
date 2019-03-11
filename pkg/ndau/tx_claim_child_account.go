package ndau

import (
	"bytes"
	"fmt"

	metast "github.com/oneiro-ndev/metanode/pkg/meta/state"
	"github.com/oneiro-ndev/ndau/pkg/ndau/backing"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	"github.com/oneiro-ndev/ndaumath/pkg/signature"
	sv "github.com/oneiro-ndev/system_vars/pkg/system_vars"
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

	// The child signature should properly sign the address bytes using the child ownership key.
	if !tx.ChildSignature.Verify([]byte(tx.Child.String()), tx.ChildOwnership) {
		return errors.New("Invalid child signature")
	}

	// Similar to ClaimAccount tx: there must be at least 1 and no more than a const transfer
	// keys set in this tx.
	if len(tx.ChildValidationKeys) < 1 || len(tx.ChildValidationKeys) > backing.MaxKeysInAccount {
		return fmt.Errorf("Expect between 1 and %d transfer keys; got %d",
			backing.MaxKeysInAccount, len(tx.ChildValidationKeys))
	}

	// No child transfer key may be equal to the child ownership key.
	for _, tk := range tx.ChildValidationKeys {
		if bytes.Equal(tk.KeyBytes(), tx.ChildOwnership.KeyBytes()) {
			return errors.New("Child ownership key may not be used as a transfer key")
		}
	}

	// Ensure the validation scripts are chaincode.
	if len(tx.ChildValidationScript) > 0 && !IsChaincode(tx.ChildValidationScript) {
		return errors.New("Child validation script must be chaincode")
	}

	app := appI.(*App)

	_, _, _, err = app.getTxAccount(tx)
	if err != nil {
		return err
	}

	child, _ := app.getAccount(tx.Child)

	// Ensure the child account is not already claimed.
	// This consequently also ensures that we won't have ancestry loops.
	// It also ensures that the child account does not yet have a Parent or Progenitor set,
	// as that happens in Apply().  Those are set precisely when its validation keys are set.
	// So only checking validation keys here is sufficient.
	if len(child.ValidationKeys) > 0 {
		return errors.New("Cannot claim a child account that is already claimed")
	}

	// Below we ensure that the child account is not locked.
	// This rule only applies for exchange accounts.

	// We look at the target (i.e. parent) account for whether the child is (rather, will be) an
	// exchange account.  The parent account is fully established at this point; the child account
	// may not be.  In Apply(), the child account will be "adopted" and automatically inherit all
	// attributes from the parent when the child's Progenitor is set.
	isExchangeAccount, err := app.accountHasAttribute(tx.Target, sv.AccountAttributeExchange)
	if err != nil {
		return err
	}

	// Here we can check the child account state to see if it's locked.  It's possible that it
	// already exists, has a balance, is locked, etc.  If it doesn't exist, the Lock field will
	// be nil, and therefore unlocked by default.  If it does exist, the Lock field will be
	// non-nil and we have to look closer at the notification state before determining whether
	// it's currently locked.
	if isExchangeAccount && child.IsLocked(app.blockTime) {
		return errors.New("Cannot claim a locked child exchange account")
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

		period := tx.ChildSettlementPeriod
		if period < 0 {
			period = app.getDefaultSettlementDuration()
		}
		child.SettlementSettings.Period = period

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

// ExtendSignatures implements Signable
func (tx *ClaimChildAccount) ExtendSignatures(sa []signature.Signature) {
	tx.Signatures = append(tx.Signatures, sa...)
}
