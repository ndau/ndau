package ndau

import (
	"fmt"

	metast "github.com/oneiro-ndev/metanode/pkg/meta/state"
	"github.com/oneiro-ndev/ndau/pkg/ndau/backing"
	sv "github.com/oneiro-ndev/ndau/pkg/ndau/system_vars"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	math "github.com/oneiro-ndev/ndaumath/pkg/types"
	"github.com/oneiro-ndev/signature/pkg/signature"
)

// NewClaimNodeReward creates a new ClaimNodeReward transaction
func NewClaimNodeReward(node address.Address, sequence uint64, keys []signature.PrivateKey) *ClaimNodeReward {
	tx := &ClaimNodeReward{Node: node, Sequence: sequence}
	for _, key := range keys {
		tx.Signatures = append(tx.Signatures, key.Sign(tx.SignableBytes()))
	}
	return tx
}

// SignableBytes implements Transactable
func (tx *ClaimNodeReward) SignableBytes() []byte {
	bytes := make([]byte, 0, 8+len(tx.Node.String()))
	bytes = appendUint64(bytes, tx.Sequence)
	bytes = append(bytes, tx.Node.String()...)
	return bytes
}

// Validate implements metatx.Transactable
func (tx *ClaimNodeReward) Validate(appI interface{}) error {
	app := appI.(*App)

	_, _, _, err := app.getTxAccount(tx)
	if err != nil {
		return err
	}

	timeout := math.Duration(0)
	err = app.System(sv.NodeRewardNominationTimeoutName, &timeout)
	if err != nil {
		return err
	}

	state := app.GetState().(*backing.State)

	if app.blockTime.Compare(state.LastNodeRewardNomination.Add(timeout)) > 0 {
		return fmt.Errorf(
			"too late: NominateNodeReward @ %s expired after %s; currently %s",
			state.LastNodeRewardNomination,
			timeout,
			app.blockTime,
		)
	}

	return nil
}

// Apply implements metatx.Transactable
func (tx *ClaimNodeReward) Apply(appI interface{}) error {
	app := appI.(*App)
	err := app.applyTxDetails(tx)
	if err != nil {
		return err
	}

	return app.UpdateState(func(stateI metast.State) (metast.State, error) {
		state := stateI.(*backing.State)

		// TODO: run the node distribution script

		return state, nil
	})
}

// GetSource implements sourcer
func (tx *ClaimNodeReward) GetSource(*App) (address.Address, error) {
	return tx.Node, nil
}

// GetSequence implements sequencer
func (tx *ClaimNodeReward) GetSequence() uint64 {
	return tx.Sequence
}

// GetSignatures implements signeder
func (tx *ClaimNodeReward) GetSignatures() []signature.Signature {
	return tx.Signatures
}
