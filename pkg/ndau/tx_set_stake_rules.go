package ndau

import (
	"fmt"

	metast "github.com/oneiro-ndev/metanode/pkg/meta/state"
	"github.com/oneiro-ndev/ndau/pkg/ndau/backing"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	"github.com/oneiro-ndev/ndaumath/pkg/signature"
	"github.com/pkg/errors"
)

// Validate implements metatx.Transactable
func (tx *SetStakeRules) Validate(appI interface{}) (err error) {
	tx.Target, err = address.Validate(tx.Target.String())
	if err != nil {
		return
	}

	if len(tx.StakeRules) > 0 && !IsChaincode(tx.StakeRules) {
		return errors.New("Stake rules must be chaincode")
	}

	app := appI.(*App)
	ad, _, _, err := app.getTxAccount(tx)
	if err != nil {
		return err
	}

	if ad.StakeRules != nil && len(ad.StakeRules.Inbound) > 0 {
		return fmt.Errorf(
			"cannot change stake rules: %d accounts are staked to this rules account",
			len(ad.StakeRules.Inbound))
	}

	return
}

// Apply implements metatx.Transactable
func (tx *SetStakeRules) Apply(appI interface{}) error {
	app := appI.(*App)

	return app.UpdateState(app.applyTxDetails(tx), func(stateI metast.State) (metast.State, error) {
		state := stateI.(*backing.State)
		ad, _ := app.getAccount(tx.Target)

		ad.StakeRules = &backing.StakeRules{
			Script:  tx.StakeRules,
			Inbound: make(map[string]uint64),
		}
		if len(tx.StakeRules) == 0 {
			ad.StakeRules = nil
		}

		state.Accounts[tx.Target.String()] = ad
		return state, nil
	})
}

// GetSource implements Sourcer
func (tx *SetStakeRules) GetSource(*App) (address.Address, error) {
	return tx.Target, nil
}

// GetSequence implements Sequencer
func (tx *SetStakeRules) GetSequence() uint64 {
	return tx.Sequence
}

// GetSignatures implements Signeder
func (tx *SetStakeRules) GetSignatures() []signature.Signature {
	return tx.Signatures
}

// ExtendSignatures implements Signable
func (tx *SetStakeRules) ExtendSignatures(sa []signature.Signature) {
	tx.Signatures = append(tx.Signatures, sa...)
}
