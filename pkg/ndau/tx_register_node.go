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
	"fmt"

	metast "github.com/ndau/metanode/pkg/meta/state"
	"github.com/ndau/ndau/pkg/ndau/backing"
	"github.com/ndau/ndaumath/pkg/address"
	"github.com/ndau/ndaumath/pkg/signature"
	sv "github.com/ndau/system_vars/pkg/system_vars"
	"github.com/pkg/errors"
	tc "github.com/tendermint/tendermint/crypto"
	ted "github.com/tendermint/tendermint/crypto/ed25519"
	tsecp "github.com/tendermint/tendermint/crypto/secp256k1"
)

// TMAddress constructs a Tendermint-style address from an ndau format public key
func TMAddress(key signature.PublicKey) (string, error) {
	var tkey tc.PubKey
	switch {
	case signature.SameAlgorithm(key.Algorithm(), signature.Ed25519):
		var data [ted.PubKeyEd25519Size]byte
		copy(data[:], key.KeyBytes())
		tkey = ted.PubKeyEd25519(data)
	case signature.SameAlgorithm(key.Algorithm(), signature.Secp256k1):
		var data [tsecp.PubKeySecp256k1Size]byte
		copy(data[:], key.KeyBytes())
		tkey = tsecp.PubKeySecp256k1(data)
	default:
		return "", errors.New("unknown key algorithm")
	}

	return tkey.Address().String(), nil
}

// Validate implements metatx.Transactable
func (tx *RegisterNode) Validate(appI interface{}) error {
	if !IsChaincode(tx.DistributionScript) {
		return errors.New("DistributionScript invalid")
	}

	app := appI.(*App)
	state := app.GetState().(*backing.State)

	target, _, _, err := app.getTxAccount(tx)
	if err != nil {
		return err
	}

	var noderules address.Address
	err = app.System(sv.NodeRulesAccountAddressName, &noderules)
	if err != nil {
		return errors.Wrap(err, "getting node rules sysvar")
	}

	if target.PrimaryStake(noderules) == nil {
		return errors.New("target must be primary staker to node rules account")
	}

	genAddr, err := address.Generate(tx.Node.Kind(), tx.Ownership.KeyBytes())
	if err != nil {
		return errors.Wrap(err, "generating address for validation")
	}
	if genAddr != tx.Node {
		return errors.New("ownership key did not generate node address")
	}

	// node key must be ed25519 for now because that's the only kind we
	// currently know how to generate validator updates for
	if !signature.SameAlgorithm(signature.Ed25519, tx.Ownership.Algorithm()) {
		return errors.New("node ownership keys must be ed25519")
	}

	_, err = TMAddress(tx.Ownership)
	if err != nil {
		return errors.Wrap(err, "generating TM address from ownership key")
	}

	if state.IsActiveNode(tx.Node) {
		return errors.New("node is already active")
	}

	vm, err := BuildVMForRulesValidation(tx, state, noderules)
	if err != nil {
		return errors.Wrap(err, "building rules validation vm")
	}
	err = vm.Run(nil)
	if err != nil {
		return errors.Wrap(err, "running rules validation vm")
	}
	returncode, err := vm.Stack().PopAsInt64()
	if err != nil {
		return errors.Wrap(err, "getting return code from rules validation vm")
	}
	if returncode != 0 {
		return fmt.Errorf("rules validation script returned code %d", returncode)
	}

	return nil
}

func (app *App) registerNode(
	nodeA address.Address,
	distributionScript []byte,
	ownership signature.PublicKey,
) func(stateI metast.State) (metast.State, error) {
	return func(stateI metast.State) (metast.State, error) {
		tma, err := TMAddress(ownership)
		if err != nil {
			return stateI, errors.Wrap(err, "generating TM address from ownership key")
		}

		state := stateI.(*backing.State)
		node := state.Nodes[nodeA.String()]

		node.Active = true
		node.DistributionScript = distributionScript
		node.TMAddress = tma
		node.Key = ownership

		if app.IsFeatureActive("NodeRegistrationDate") {
			node.SetRegistration(app.BlockTime())
		}

		state.Nodes[nodeA.String()] = node

		return state, nil
	}
}

// Apply implements metatx.Transactable
func (tx *RegisterNode) Apply(appI interface{}) error {
	app := appI.(*App)
	return app.UpdateState(
		app.applyTxDetails(tx),
		app.registerNode(tx.Node, tx.DistributionScript, tx.Ownership))
}

// GetSource implements Sourcer
func (tx *RegisterNode) GetSource(*App) (address.Address, error) {
	return tx.Node, nil
}

// GetSequence implements Sequencer
func (tx *RegisterNode) GetSequence() uint64 {
	return tx.Sequence
}

// GetSignatures implements Signeder
func (tx *RegisterNode) GetSignatures() []signature.Signature {
	return tx.Signatures
}

// ExtendSignatures implements Signable
func (tx *RegisterNode) ExtendSignatures(sa []signature.Signature) {
	tx.Signatures = append(tx.Signatures, sa...)
}
