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
func (tx *GTValidatorChange) Validate(interface{}) error {
	return nil
}

// ToValidator converts this struct into a ValidatorUpdate
func (tx *GTValidatorChange) ToValidator() abci.ValidatorUpdate {
	return abci.Ed25519ValidatorUpdate(tx.PublicKey, tx.Power)
}

// Apply this GTVC to the node state
func (tx *GTValidatorChange) Apply(appInt interface{}) error {
	app := appInt.(*App)

	// from persistent_app.go: we now know that this public key should be in go-crypto format
	logger := app.GetLogger().WithField("method", "GTValidatorChange.Apply")
	logger.WithFields(log.Fields{
		"PubKey": fmt.Sprintf("%x", tx.PublicKey),
		"Power":  tx.Power,
	}).Info("entered method")
	if err := tx.Validate(app); err != nil {
		logger.Info("exit method; invalid tx")
		return err
	}
	v := tx.ToValidator()
	app.UpdateValidator(v)

	logger.Info("exit method; success")
	return nil
}

// SignableBytes implements Transactable
func (tx *GTValidatorChange) SignableBytes() []byte {
	bytes, err := tx.MarshalMsg(nil)
	panicIfError(err, "GTVC signable bytes non nil error")
	return bytes
}

func panicIfError(err error, msg string) {
	if err != nil {
		panic(msg)
	}
}
