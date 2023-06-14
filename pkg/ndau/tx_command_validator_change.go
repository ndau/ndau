package ndau

// ----- ---- --- -- -
// Copyright 2019 Oneiro NA, Inc. All Rights Reserved.
//
// Licensed under the Apache License 2.0 (the "License").  You may not use
// this file except in compliance with the License.  You can obtain a copy
// in the file LICENSE in the source distribution or at
// https://www.apache.org/licenses/LICENSE-2.0.txt
// - -- --- ---- -----

import (
	metatx "github.com/ndau/metanode/pkg/meta/transaction"
	"github.com/ndau/msgp-well-known-types/wkt"
	"github.com/ndau/ndau/pkg/ndau/backing"
	"github.com/ndau/ndaumath/pkg/address"
	"github.com/ndau/ndaumath/pkg/signature"
	sv "github.com/ndau/system_vars/pkg/system_vars"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	abci "github.com/oneiro-ndev/tendermint.0.32.3/abci/types"
)

// Validate implements metatx.Transactable
func (tx *CommandValidatorChange) Validate(appI interface{}) error {
	app := appI.(*App)

	var maxValidators wkt.Uint64
	err := app.System(sv.NodeMaxValidators, &maxValidators)
	if err == nil {
		return errors.New("CVC disallowed when MAX_VALIDATORS is set")
	}

	err = tx.Node.Revalidate()
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
			"method":     "CommandValidatorChange.Validate",
			"txHash":     metatx.Hash(tx),
			"acctExists": exists,
			"signatures": sigs,
		})
		logger.Error("cvc validation err")
	}

	return err
}

// Apply this CVC to the node state
func (tx *CommandValidatorChange) Apply(appI interface{}) error {
	app := appI.(*App)
	err := app.UpdateState(app.applyTxDetails(tx))
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

// create an abci.ValidatorUpdate with 0 power for a given node
func validatorUpdateFor(state *backing.State, node string) (*abci.ValidatorUpdate, error) {
	n, ok := state.Nodes[node]
	if !ok || !n.Active {
		return nil, errors.New("node must be active")
	}
	if !signature.SameAlgorithm(n.Key.Algorithm(), signature.Ed25519) {
		return nil, errors.New("node key must be an Ed25519")
	}
	vu := abci.Ed25519ValidatorUpdate(n.Key.KeyBytes(), 0)
	return &vu, nil
}

// ToValidator converts this tx into a TM-style ValidatorUpdate struct
func (tx *CommandValidatorChange) ToValidator(state *backing.State) (*abci.ValidatorUpdate, error) {
	vu, err := validatorUpdateFor(state, tx.Node.String())
	vu.Power = tx.Power
	return vu, err
}

// ExtendSignatures implements Signable
func (tx *CommandValidatorChange) ExtendSignatures(sa []signature.Signature) {
	tx.Signatures = append(tx.Signatures, sa...)
}
