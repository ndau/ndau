package ndau

import (
	uuid "github.com/satori/go.uuid"
	"github.com/tendermint/abci/types"
)

// static assert that GTValidatorChange fits the Transactable
// interface
var _ Transactable = (*GTValidatorChange)(nil)

// IsValid returns nil if tx is valid, or an error otherwise
func (tx *GTValidatorChange) IsValid(app *App) error {
	return nil
}

// Apply applies this GTVC to the supplied app
func (tx *GTValidatorChange) Apply(app *App) error {
	if err := tx.IsValid(app); err != nil {
		return err
	}
	v := tx.ToValidator()

	app.updateValidator(v)
	return nil
}

// AsTransaction builds a Transaction from a GTValidatorChange
func (tx *GTValidatorChange) AsTransaction() *Transaction {
	return &Transaction{
		Tx: &Transaction_GtValidatorChange{
			GtValidatorChange: tx,
		},
		Nonce: uuid.NewV1().Bytes(),
	}
}

// ToValidator converts this struct into a Validator
func (tx *GTValidatorChange) ToValidator() types.Validator {
	return types.Validator{
		PubKey: tx.PublicKey,
		Power:  tx.Power,
	}
}
