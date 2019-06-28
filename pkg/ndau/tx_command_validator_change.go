package ndau

import (
	metatx "github.com/oneiro-ndev/metanode/pkg/meta/transaction"
	"github.com/oneiro-ndev/ndau/pkg/ndau/backing"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	"github.com/oneiro-ndev/ndaumath/pkg/signature"
	sv "github.com/oneiro-ndev/system_vars/pkg/system_vars"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	abci "github.com/tendermint/tendermint/abci/types"
)

// Validate implements metatx.Transactable
func (tx *CommandValidatorChange) Validate(appI interface{}) error {
	app := appI.(*App)

	err := tx.Node.Revalidate()
	if err != nil {
		return errors.Wrap(err, "node address")
	}

	state := app.GetState().(*backing.State)
	node, ok := state.Nodes[tx.Node.String()]
	if !ok || !node.Active {
		return errors.New("node must be active")
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
	_, err := app.applyTxDetails(tx)(app.GetState())
	if err != nil {
		return err
	}

	// unusually, we don't actually directly touch app state in this tx
	// instead, we call UpdateValidator, which updates the metastate
	// appropriately.
	vu, err := tx.ToValidator(app.GetState().(*backing.State))
	if err != nil {
		return errors.Wrap(err, "constructing TM ValidatorUpdate")
	}
	app.UpdateValidator(*vu)
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

// ToValidator converts this tx into a TM-style ValidatorUpdate struct
func (tx *CommandValidatorChange) ToValidator(state *backing.State) (*abci.ValidatorUpdate, error) {
	node, ok := state.Nodes[tx.Node.String()]
	if !ok || !node.Active {
		return nil, errors.New("node must be active")
	}
	if !signature.SameAlgorithm(node.Key.Algorithm(), signature.Ed25519) {
		return nil, errors.New("node key must be an Ed25519")
	}
	vu := abci.Ed25519ValidatorUpdate(node.Key.KeyBytes(), tx.Power)
	return &vu, nil
}

// ExtendSignatures implements Signable
func (tx *CommandValidatorChange) ExtendSignatures(sa []signature.Signature) {
	tx.Signatures = append(tx.Signatures, sa...)
}
