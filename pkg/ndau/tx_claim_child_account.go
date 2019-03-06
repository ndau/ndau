package ndau

import (
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	"github.com/oneiro-ndev/ndaumath/pkg/signature"
)

// GetAccountAddresses returns the account addresses associated with this transaction type.
func (tx *ClaimChildAccount) GetAccountAddresses() []string {
	return []string{tx.Target.String(), tx.Child.String()}
}

// Validate returns nil if tx is valid, or an error
func (tx *ClaimChildAccount) Validate(appI interface{}) error {
	// TODO: Implement.
	return nil
}

// Apply applies this tx if no error occurs
func (tx *ClaimChildAccount) Apply(appI interface{}) error {
	// TODO: Implement.
	return nil
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
