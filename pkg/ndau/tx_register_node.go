package ndau

import (
	"fmt"

	metast "github.com/oneiro-ndev/metanode/pkg/meta/state"
	"github.com/oneiro-ndev/ndau/pkg/ndau/backing"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	"github.com/oneiro-ndev/ndaumath/pkg/signature"
	sv "github.com/oneiro-ndev/system_vars/pkg/system_vars"
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

// GetAccountAddresses returns the account addresses associated with this transaction type.
func (tx *RegisterNode) GetAccountAddresses() []string {
	return []string{tx.Node.String()}
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

// Apply implements metatx.Transactable
func (tx *RegisterNode) Apply(appI interface{}) error {
	app := appI.(*App)
	err := app.applyTxDetails(tx)
	if err != nil {
		return err
	}
	tma, err := TMAddress(tx.Ownership)
	if err != nil {
		return errors.Wrap(err, "generating TM address from ownership key")
	}

	return app.UpdateState(func(stateI metast.State) (metast.State, error) {
		state := stateI.(*backing.State)
		node := state.Nodes[tx.Node.String()]

		node.Active = true
		node.DistributionScript = tx.DistributionScript
		node.TMAddress = tma

		state.Nodes[tx.Node.String()] = node

		return state, nil
	})
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
