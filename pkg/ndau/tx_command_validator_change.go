package ndau

import (
	"fmt"

	sv "github.com/oneiro-ndev/ndau/pkg/ndau/system_vars"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	"github.com/oneiro-ndev/ndaumath/pkg/signature"
	log "github.com/sirupsen/logrus"
	abci "github.com/tendermint/tendermint/abci/types"
)

// SignableBytes implements Transactable
func (tx *CommandValidatorChange) SignableBytes() []byte {
	bytes := make([]byte, len(tx.PublicKey), 8+8+len(tx.PublicKey))
	bytes = appendUint64(bytes, uint64(tx.Power))
	bytes = appendUint64(bytes, tx.Sequence)
	return bytes
}

// NewCommandValidatorChange constructs a CommandValidatorChange transactable.
func NewCommandValidatorChange(
	publicKey []byte,
	power int64,
	sequence uint64,
	keys []signature.PrivateKey,
) (tx CommandValidatorChange) {
	tx = CommandValidatorChange{
		PublicKey: publicKey,
		Power:     power,
		Sequence:  sequence,
	}
	for _, key := range keys {
		tx.Signatures = append(tx.Signatures, key.Sign(tx.SignableBytes()))
	}
	return tx
}

// Validate implements metatx.Transactable
func (tx *CommandValidatorChange) Validate(appI interface{}) error {
	app := appI.(*App)

	_, _, _, err := app.getTxAccount(tx)
	if err != nil {
		return err
	}

	return err
}

// Apply this GTVC to the node state
func (tx *CommandValidatorChange) Apply(appInt interface{}) error {
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

// GetSource implements sourcer
func (tx *CommandValidatorChange) GetSource(app *App) (addr address.Address, err error) {
	err = app.System(sv.CommandValidatorChangeAddressName, &addr)
	return
}

// GetSequence implements sequencer
func (tx *CommandValidatorChange) GetSequence() uint64 {
	return tx.Sequence
}

// GetSignatures implements signeder
func (tx *CommandValidatorChange) GetSignatures() []signature.Signature {
	return tx.Signatures
}

// ToValidator converts this struct into a Validator
func (tx *CommandValidatorChange) ToValidator() abci.Validator {
	return abci.Ed25519Validator(tx.PublicKey, tx.Power)
}
