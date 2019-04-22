package ndau

import (
	metast "github.com/oneiro-ndev/metanode/pkg/meta/state"
	"github.com/oneiro-ndev/ndau/pkg/ndau/backing"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	"github.com/oneiro-ndev/ndaumath/pkg/signature"
	"github.com/pkg/errors"
)

// GetAccountAddresses returns the account addresses associated with this transaction type.
func (tx *SetStakeRules) GetAccountAddresses() []string {
	return []string{tx.Target.String()}
}

// Validate implements metatx.Transactable
func (tx *SetStakeRules) Validate(appI interface{}) (err error) {
	tx.Target, err = address.Validate(tx.Target.String())
	if err != nil {
		return
	}

	if len(tx.StakeRules) > 0 && !IsChaincode(tx.StakeRules) {
		return errors.New("Stake rules must be chaincode")
	}

	if len(tx.StakeRules) == 0 {
		// TODO: if any accounts are staked to these stake rules, it is not
		// allowed to remove the stake rules. However, as we haven't yet
		// implemented staking to a rules account, we can't write this feature
		// yet.
	}

	app := appI.(*App)
	_, _, _, err = app.getTxAccount(tx)
	if err != nil {
		return err
	}

	return
}

// Apply implements metatx.Transactable
func (tx *SetStakeRules) Apply(appI interface{}) error {
	app := appI.(*App)
	err := app.applyTxDetails(tx)
	if err != nil {
		return err
	}

	return app.UpdateState(func(stateI metast.State) (metast.State, error) {
		state := stateI.(*backing.State)
		ad, _ := app.getAccount(tx.Target)

		ad.StakeRules = &backing.StakeRules{
			Script:  tx.StakeRules,
			Inbound: make(map[string]struct{}),
		}
		if len(tx.StakeRules) == 0 {
			ad.StakeRules = nil
		}

		state.Accounts[tx.Target.String()] = ad
		return state, nil
	})
}

// GetSource implements sourcer
func (tx *SetStakeRules) GetSource(*App) (address.Address, error) {
	return tx.Target, nil
}

// GetSequence implements sequencer
func (tx *SetStakeRules) GetSequence() uint64 {
	return tx.Sequence
}

// GetSignatures implements signeder
func (tx *SetStakeRules) GetSignatures() []signature.Signature {
	return tx.Signatures
}

// ExtendSignatures implements Signable
func (tx *SetStakeRules) ExtendSignatures(sa []signature.Signature) {
	tx.Signatures = append(tx.Signatures, sa...)
}
