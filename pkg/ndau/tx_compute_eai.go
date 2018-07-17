package ndau

import (
	"encoding/binary"
	"fmt"

	metast "github.com/oneiro-ndev/metanode/pkg/meta/state"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	"github.com/oneiro-ndev/ndaumath/pkg/eai"
	"github.com/oneiro-ndev/ndau/pkg/ndau/backing"
	sv "github.com/oneiro-ndev/ndau/pkg/ndau/system_vars"
	"github.com/oneiro-ndev/signature/pkg/signature"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

// NewComputeEAI creates a new ComputeEAI transaction
//
// Most users will never need this.
func NewComputeEAI(node address.Address, sequence uint64, key signature.PrivateKey) *ComputeEAI {
	c := &ComputeEAI{Node: node, Sequence: sequence}
	c.Signature = key.Sign(c.signableBytes())
	return c
}

func (c *ComputeEAI) signableBytes() []byte {
	bytes := make([]byte, 8, 8+len(c.Node.String()))
	binary.BigEndian.PutUint64(bytes, c.Sequence)
	bytes = append(bytes, c.Node.String()...)
	return bytes
}

// Validate implements metatx.Transactable
func (c *ComputeEAI) Validate(appI interface{}) error {
	app := appI.(*App)
	state := app.GetState().(*backing.State)

	nodeData, hasNode := state.Accounts[c.Node.String()]
	if !hasNode {
		return errors.New("No such node")
	}
	// is the tx sequence higher than the highest previous sequence?
	if c.Sequence <= nodeData.Sequence {
		return errors.New("Sequence too low")
	}
	// does the signature check out?
	if nodeData.TransferKey == nil {
		return errors.New("Transfer key not set")
	}
	if !nodeData.TransferKey.Verify(c.signableBytes(), c.Signature) {
		return errors.New("Invalid signature")
	}

	return nil
}

// Apply implements metatx.Transactable
func (c *ComputeEAI) Apply(appI interface{}) error {
	app := appI.(*App)
	unlockedTable := new(eai.RateTable)
	err := app.System(sv.UnlockedRateTableName, unlockedTable)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("Error fetching %s system variable in ComputeEAI.Apply", sv.UnlockedRateTableName))
	}
	lockedTable := new(eai.RateTable)
	err = app.System(sv.LockedRateTableName, lockedTable)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("Error fetching %s system variable in ComputeEAI.Apply", sv.UnlockedRateTableName))
	}

	return app.UpdateState(func(stateI metast.State) (metast.State, error) {
		state := stateI.(*backing.State)
		nodeData := state.Accounts[c.Node.String()]
		nodeData.Sequence = c.Sequence
		state.Accounts[c.Node.String()] = nodeData

		delegatedAccounts := state.Delegates[c.Node.String()]
		var errorList []error
		for accountAddr := range delegatedAccounts {
			acctData, hasAcct := state.Accounts[accountAddr]
			if !hasAcct {
				// accounts can sometimes be removed, i.e. due to 0 balance
				// if we encounter that, don't worry about it
				// TODO we may want to actually remove the reference to
				// deleted accounts
				continue
			}
			logger := app.GetLogger().WithFields(log.Fields{
				"tx":            "ComputeEAI",
				"acct":          accountAddr,
				"node":          c.Node.String(),
				"acctData":      acctData,
				"blockTime":     app.blockTime,
				"unlockedTable": unlockedTable,
				"lockedTable":   lockedTable,
			})
			err = acctData.WeightedAverageAge.UpdateWeightedAverageAge(
				app.blockTime.Since(acctData.LastWAAUpdate),
				0,
				acctData.Balance,
			)
			if err != nil {
				// the only error expected from an EAI calculation is overflowing
				// the Ndau type. If that happens, ndau is broken. If ndau
				// is broken, then we don't have to uphold its promises for that
				// account. This means that we'll log the error, but then just
				// proceed as if no error had occurred.
				//
				// The other option would be to panic if things like this
				// happened, and we choose to follow Douglas Adams' advice.
				logger.WithError(err).Error("eai.Calculate failed")
				errorList = append(errorList, err)
				err = nil
				continue
			}
			acctData.LastWAAUpdate = app.blockTime

			eaiAward, err := eai.Calculate(
				acctData.Balance, app.blockTime, acctData.LastEAIUpdate,
				acctData.WeightedAverageAge, acctData.Lock,
				*unlockedTable, *lockedTable,
			)
			if err != nil {
				errorList = append(errorList, err)
				err = nil
				continue
			}
			acctData.Balance, err = acctData.Balance.Add(eaiAward)
			if err != nil {
				// same deal: we either panic, or just log the error and soldier on
				logger.WithError(err).WithField("eaiAward", eaiAward).Error("error updating account balance")
				errorList = append(errorList, err)
				err = nil
				continue
			}
			acctData.LastEAIUpdate = app.blockTime
			state.Accounts[accountAddr] = acctData
		}
		if len(errorList) > 0 {
			errStr := fmt.Sprintf("Errors found calculating EAI for node %s: ", c.Node.String())
			for idx, err := range errorList {
				if idx != 0 {
					errStr += ", "
				}
				errStr += err.Error()
			}
			err = errors.New(errStr)
		}
		return state, err
	})
}
