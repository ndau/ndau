package ndau

import (
	"fmt"

	"github.com/oneiro-ndev/metanode/pkg/meta/transaction"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	"github.com/oneiro-ndev/ndaumath/pkg/signature"
	sv "github.com/oneiro-ndev/system_vars/pkg/system_vars"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto/ed25519"
)

// Validate implements metatx.Transactable
func (tx *CommandValidatorChange) Validate(appI interface{}) error {
	app := appI.(*App)

	if len(tx.PublicKey) == 0 {
		return errors.New("cvc must have non-empty public key")
	}

	if len(tx.PublicKey) != ed25519.PubKeyEd25519Size {
		return fmt.Errorf(
			"Wrong length for Ed25519 public key: want %d, have %d",
			ed25519.PubKeyEd25519Size,
			len(tx.PublicKey),
		)
	}

	_, exists, signatures, err := app.getTxAccount(tx)
	if err != nil {
		sigs := ""
		if signatures != nil {
			sigs = signatures.String()
		}
		logger := app.GetLogger().WithError(err).WithFields(logrus.Fields{
			"method":      "CommandValidatorChange.Validate",
			"tx hash":     metatx.Hash(tx),
			"acct exists": exists,
			"signatures":  sigs,
		})
		logger.Error("cvc validation err")
	}

	return err
}

// Apply this CVC to the node state
func (tx *CommandValidatorChange) Apply(appI interface{}) error {
	app := appI.(*App)
	err := app.applyTxDetails(tx)
	if err != nil {
		return err
	}

	// unusually, we don't actually directly touch app state in this tx
	// instead, we call UpdateValidator, which updates the metastate
	// appropriately.
	app.UpdateValidator(tx.ToValidator())
	return nil
}

// GetSource implements Sourcer
func (tx *CommandValidatorChange) GetSource(app *App) (addr address.Address, err error) {
	err = app.System(sv.CommandValidatorChangeAddressName, &addr)
	return
}

// GetSequence implements Sequencer
func (tx *CommandValidatorChange) GetSequence() uint64 {
	return tx.Sequence
}

// GetSignatures implements Signeder
func (tx *CommandValidatorChange) GetSignatures() []signature.Signature {
	return tx.Signatures
}

// ToValidator converts this struct into a ValidatorUpdate
func (tx *CommandValidatorChange) ToValidator() abci.ValidatorUpdate {
	return abci.Ed25519ValidatorUpdate(tx.PublicKey, tx.Power)
}

// ExtendSignatures implements Signable
func (tx *CommandValidatorChange) ExtendSignatures(sa []signature.Signature) {
	tx.Signatures = append(tx.Signatures, sa...)
}
