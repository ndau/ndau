package ndau

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	abci "github.com/tendermint/tendermint/abci/types"
)

// Validate implements business logic determining if a GTValidatorChange is valid.
//
// As it happens, we don't currently _have_ any business logic which would
// determine whether or not a GTValidatorChange is valid, but we've put in
// this method anyway to implement the Transactable interface
func (vc *GTValidatorChange) Validate(interface{}) error {
	return nil
}

// ToValidator converts this struct into a Validator
func (vc *GTValidatorChange) ToValidator() abci.Validator {
	return abci.Ed25519Validator(vc.PublicKey, vc.Power)
}

// Apply this GTVC to the node state
func (vc *GTValidatorChange) Apply(appInt interface{}) error {
	app := appInt.(*App)

	// from persistent_app.go: we now know that this public key should be in go-crypto format
	logger := app.GetLogger().WithField("method", "GTValidatorChange.Apply")
	logger.WithFields(log.Fields{
		"PubKey": fmt.Sprintf("%x", vc.PublicKey),
		"Power":  vc.Power,
	}).Info("entered method")
	if err := vc.Validate(app); err != nil {
		logger.Info("exit method; invalid vc")
		return err
	}
	v := vc.ToValidator()
	app.UpdateValidator(v)

	logger.Info("exit method; success")
	return nil
}

func panicIfError(err error, msg string) {
	if err != nil {
		panic(msg)
	}
}
