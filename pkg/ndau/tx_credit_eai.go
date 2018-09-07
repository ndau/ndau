package ndau

import (
	"fmt"

	metast "github.com/oneiro-ndev/metanode/pkg/meta/state"
	"github.com/oneiro-ndev/ndau/pkg/ndau/backing"
	sv "github.com/oneiro-ndev/ndau/pkg/ndau/system_vars"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	"github.com/oneiro-ndev/ndaumath/pkg/eai"
	"github.com/oneiro-ndev/signature/pkg/signature"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

// NewCreditEAI creates a new CreditEAI transaction
//
// Most users will never need this.
func NewCreditEAI(node address.Address, sequence uint64, keys []signature.PrivateKey) *CreditEAI {
	tx := &CreditEAI{Node: node, Sequence: sequence}
	for _, key := range keys {
		tx.Signatures = append(tx.Signatures, key.Sign(tx.SignableBytes()))
	}
	return tx
}

// SignableBytes implements Transactable
func (tx *CreditEAI) SignableBytes() []byte {
	bytes := make([]byte, 0, 8+len(tx.Node.String()))
	bytes = appendUint64(bytes, tx.Sequence)
	bytes = append(bytes, tx.Node.String()...)
	return bytes
}

// Validate implements metatx.Transactable
func (tx *CreditEAI) Validate(appI interface{}) error {
	app := appI.(*App)

	_, hasNode, _, err := app.getTxAccount(
		tx,
		tx.Node,
		tx.Sequence,
		tx.Signatures,
	)
	if err != nil {
		return err
	}

	if !hasNode {
		return errors.New("No such node")
	}

	return nil
}

// Apply implements metatx.Transactable
func (tx *CreditEAI) Apply(appI interface{}) error {
	app := appI.(*App)
	unlockedTable := new(eai.RateTable)
	err := app.System(sv.UnlockedRateTableName, unlockedTable)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("Error fetching %s system variable in CreditEAI.Apply", sv.UnlockedRateTableName))
	}
	lockedTable := new(eai.RateTable)
	err = app.System(sv.LockedRateTableName, lockedTable)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("Error fetching %s system variable in CreditEAI.Apply", sv.UnlockedRateTableName))
	}

	return app.UpdateState(func(stateI metast.State) (metast.State, error) {
		state := stateI.(*backing.State)
		nodeData, _ := state.GetAccount(tx.Node, app.blockTime)
		nodeData.Sequence = tx.Sequence

		fee, err := app.calculateTxFee(tx)
		if err != nil {
			return state, err
		}
		nodeData.Balance -= fee
		state.Accounts[tx.Node.String()] = nodeData

		delegatedAccounts := state.Delegates[tx.Node.String()]
		var errorList []error
		for accountAddrS := range delegatedAccounts {
			accountAddr, err := address.Validate(accountAddrS)
			if err != nil {
				return state, errors.Wrap(err, "CreditEAI: validating delegated account address")
			}
			acctData, hasAcct := state.GetAccount(accountAddr, app.blockTime)
			if !hasAcct {
				// accounts can sometimes be removed, i.e. due to 0 balance
				// if we encounter that, don't worry about it
				// TODO we may want to actually remove the reference to
				// deleted accounts
				continue
			}
			logger := app.GetLogger().WithFields(log.Fields{
				"tx":            "CreditEAI",
				"acct":          accountAddr,
				"node":          tx.Node.String(),
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

			destAcct := accountAddr
			destAcctData := acctData
			if acctData.RewardsTarget != nil {
				destAcct = *acctData.RewardsTarget
				destAcctData, _ = state.GetAccount(destAcct, app.blockTime)
			}

			destAcctData.Balance, err = destAcctData.Balance.Add(eaiAward)
			if err != nil {
				// same deal: we either panic, or just log the error and soldier on
				logger.WithError(err).WithField("eaiAward", eaiAward).Error("error updating account balance")
				errorList = append(errorList, err)
				err = nil
				continue
			}
			acctData.LastEAIUpdate = app.blockTime

			if destAcct.String() == accountAddr.String() {
				// special case: just update the acctData balance, and
				// make a single state update
				acctData.Balance = destAcctData.Balance
			} else {
				state.Accounts[destAcct.String()] = destAcctData
			}
			state.Accounts[accountAddr.String()] = acctData
		}
		if len(errorList) > 0 {
			errStr := fmt.Sprintf("Errors found calculating EAI for node %s: ", tx.Node.String())
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

// GetSource implements sourcer
func (tx *CreditEAI) GetSource(*App) (address.Address, error) {
	return tx.Node, nil
}

// GetSequence implements sequencer
func (tx *CreditEAI) GetSequence() uint64 {
	return tx.Sequence
}
