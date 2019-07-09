package ndau

import (
	"bytes"
	"fmt"

	metast "github.com/oneiro-ndev/metanode/pkg/meta/state"
	"github.com/oneiro-ndev/ndau/pkg/ndau/backing"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	sv "github.com/oneiro-ndev/system_vars/pkg/system_vars"
	"github.com/pkg/errors"
)

// GetAccountAddresses returns the account addresses associated with this transaction type.
func (tx *SetValidation) GetAccountAddresses() []string {
	return []string{tx.Target.String()}
}

// Validate returns nil if tx is valid, or an error
func (tx *SetValidation) Validate(appI interface{}) error {
	// we need to verify that the ownership key submitted actually generates
	// the address for which validation is being set
	// get the address kind:
	_, err := address.Validate(tx.Target.String())
	if err != nil {
		return errors.Wrap(err, "Account address invalid")
	}
	kind := tx.Target.Kind()
	if !address.IsValidKind(kind) {
		return fmt.Errorf("Account has invalid address kind: %x", kind)
	}
	ownershipAddress, err := address.Generate(kind, tx.Ownership.KeyBytes())
	if err != nil {
		return errors.Wrap(err, "generating address for ownership key")
	}

	if tx.Target.String() != ownershipAddress.String() {
		return errors.New("Ownership key and address do not match")
	}

	if !tx.Signature.Verify(tx.SignableBytes(), tx.Ownership) {
		return errors.New("Invalid ownership signature")
	}

	// business rule: there must be at least 1 and no more than a const
	// validation keys set in this tx
	if len(tx.ValidationKeys) < 1 || len(tx.ValidationKeys) > backing.MaxKeysInAccount {
		return fmt.Errorf("Expect between 1 and %d validation keys; got %d", backing.MaxKeysInAccount, len(tx.ValidationKeys))
	}

	// no validation key may be equal to the ownership key
	for _, tk := range tx.ValidationKeys {
		if bytes.Equal(tk.KeyBytes(), tx.Ownership.KeyBytes()) {
			return errors.New("Ownership key may not be used as a validation key")
		}
	}

	if len(tx.ValidationScript) > 0 && !IsChaincode(tx.ValidationScript) {
		return errors.New("Validation script must be chaincode")
	}

	app := appI.(*App)

	acct, _, _, err := app.getTxAccount(tx)
	if err != nil {
		return err
	}

	maxKeys := 1
	if app.IsFeatureActive("NoKeysOnSetValidation") {
		maxKeys = 0
	}
	if len(acct.ValidationKeys) > maxKeys {
		return fmt.Errorf("SetValidation only works when at most %d validation keys are set and %d are present",
			maxKeys, len(acct.ValidationKeys))
	}

	// Prevent establishing locked exchange accounts.  If this fails, it means we've been working
	// with an exchange, they have an address, they transferred into and locked their account, then
	// tried to create it. That will make it impossible to establish the child account: they'll have to
	// create a new address for their progenitor account.  If this check passes for a to-become child
	// exchange account, then the CreateChildAccount tx will later fail, and likewise will have to
	// make a new child account that isn't locked before establishing it as a child.
	isExchangeAccount, err := app.GetState().(*backing.State).AccountHasAttribute(tx.Target, sv.AccountAttributeExchange)
	if err != nil {
		return err
	}
	if isExchangeAccount && acct.IsLocked(app.BlockTime()) {
		return errors.New("Cannot create a locked child exchange account")
	}

	return nil
}

// Apply applies this tx if no error occurs
func (tx *SetValidation) Apply(appI interface{}) error {
	app := appI.(*App)

	return app.UpdateState(app.applyTxDetails(tx), func(stI metast.State) (metast.State, error) {
		st := stI.(*backing.State)

		acct, _ := app.getAccount(tx.Target)
		acct.ValidationKeys = tx.ValidationKeys
		acct.ValidationScript = tx.ValidationScript

		st.Accounts[tx.Target.String()] = acct

		return st, nil
	})
}

// GetSource implements Sourcer
func (tx *SetValidation) GetSource(*App) (address.Address, error) {
	return tx.Target, nil
}

// GetSequence implements Sequencer
func (tx *SetValidation) GetSequence() uint64 {
	return tx.Sequence
}
