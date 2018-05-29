package ndau

import (
	"fmt"

	abci "github.com/tendermint/abci/types"
)

// IsValid implements business logic determining if a GTValidatorChange is valid.
//
// As it happens, we don't currently _have_ any business logic which would
// determine whether or not a GTValidatorChange is valid, but we've put in
// this method anyway to implement the Transactable interface
func (vc *GTValidatorChange) IsValid(abci.Application) error {
	return nil
}

// ToValidator converts this struct into a Validator
func (vc *GTValidatorChange) ToValidator() abci.Validator {
	return abci.Validator{
		PubKey: vc.PublicKey,
		Power:  vc.Power,
	}
}

// Apply this GTVC to the node state
func (vc *GTValidatorChange) Apply(appInt abci.Application) error {
	app := appInt.(*App)

	// from persistent_app.go: we now know that this public key should be in go-crypto format
	logger := app.logger.With("method", "processGTValidatorChange")
	logger.Info(
		"entered method",
		"PubKey", fmt.Sprintf("%x", vc.PublicKey),
		"Power", vc.Power,
	)
	if err := vc.IsValid(app); err != nil {
		logger.Info("exit method; invalid vc")
		return err
	}
	v := vc.ToValidator()
	app.updateValidator(v)

	logger.Info("exit method; success")
	return nil
}

func panicIfError(err error, msg string) {
	if err != nil {
		panic(msg)
	}
}
