package ndau

import (
	"fmt"

	"github.com/oneiro-ndev/chaincode/pkg/vm"
	metast "github.com/oneiro-ndev/metanode/pkg/meta/state"
	"github.com/oneiro-ndev/ndau/pkg/ndau/backing"
	sv "github.com/oneiro-ndev/ndau/pkg/ndau/system_vars"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	math "github.com/oneiro-ndev/ndaumath/pkg/types"
	"github.com/oneiro-ndev/signature/pkg/signature"
	"github.com/pkg/errors"
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

		avm, err := BuildVMForNodeDistribution(
			state.Nodes[tx.Node.String()].DistributionScript,
			tx.Node,
			state.Nodes[tx.Node.String()].Costakers,
			state.Accounts,
			state.UnclaimedNodeReward,
			app.blockTime,
		)
		if err != nil {
			return state, errors.Wrap(err, "constructing chaincode vm")
		}
		err = avm.Run(false)
		if err != nil {
			return state, errors.Wrap(err, "running chaincode vm")
		}
		// expected top of stack on successful chaincode completion:
		// list of structs, whose field 10 is the amount returned
		// note: this list may be sorted in arbitrary order, and sort order
		// may differ from our input list.
		rv, err := avm.Stack().PopAsListOfStructs(10)
		if err != nil {
			return state, errors.Wrap(err, "pop chaincode return value as list of structs")
		}

		for _, v := range rv {
			s := v.(*vm.Struct)   // no err handling because PALOS promised
			value, _ := s.Get(10) // PALOS already checked that the field exists
			nV, isNum := value.(*vm.Number)
			if !isNum {
				return state, errors.Wrap(err, "chaincode returned non-numeric value for costaker disbursal")
			}
			n := nV.AsInt64()
			if n < 0 {
				return state, errors.New("chaincode attempted to return negative reward for costaker")
			}
			addrV, err := s.Get(1) // addresses are in field 1
			if err != nil {
				return state, errors.Wrap(err, "getting address from chaincode return value")
			}
			addrB, isBytes := addrV.(*vm.Bytes)
			if !isBytes {
				return state, errors.Wrap(err, "chaincode changed address type")
			}
			addr := addrB.String()
			addrA, err := address.Validate(addr)
			if err != nil {
				return state, errors.Wrap(err, "chaincode messed up costaker address")
			}

			// awards are disbursed first-come, first-serve to the list of
			// costakers returned by the disbursement script.
			// If the script attempts to disburse more than the actual
			// total reward, rewards are truncated and subsequent costakers
			// are just out of luck.
			award := math.Ndau(0)
			final := false
			if math.Ndau(n) <= state.UnclaimedNodeReward {
				award = math.Ndau(n)
			} else {
				award = state.UnclaimedNodeReward
				final = true
			}
			state.UnclaimedNodeReward -= award
			acct, _ := state.GetAccount(addrA, app.blockTime)
			acct.Balance += math.Ndau(n)
			state.Accounts[addr] = acct
			if final {
				break
			}
		}

		// if after disbursement to costakers there remains some node reward,
		// it goes to the node
		acct, _ := state.GetAccount(tx.Node, app.blockTime)
		acct.Balance += state.UnclaimedNodeReward
		state.Accounts[tx.Node.String()] = acct
		state.UnclaimedNodeReward = math.Ndau(0)

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
