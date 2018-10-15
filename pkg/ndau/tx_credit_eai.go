package ndau

import (
	"fmt"

	metast "github.com/oneiro-ndev/metanode/pkg/meta/state"
	"github.com/oneiro-ndev/ndau/pkg/ndau/backing"
	sv "github.com/oneiro-ndev/ndau/pkg/ndau/system_vars"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	"github.com/oneiro-ndev/ndaumath/pkg/constants"
	"github.com/oneiro-ndev/ndaumath/pkg/eai"
	"github.com/oneiro-ndev/ndaumath/pkg/signed"
	math "github.com/oneiro-ndev/ndaumath/pkg/types"
	"github.com/oneiro-ndev/ndaumath/pkg/signature"
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

	_, hasNode, _, err := app.getTxAccount(tx)
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
	err := app.applyTxDetails(tx)
	if err != nil {
		return err
	}

	unlockedTable := new(eai.RateTable)
	err = app.System(sv.UnlockedRateTableName, unlockedTable)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("Error fetching %s system variable in CreditEAI.Apply", sv.UnlockedRateTableName))
	}
	lockedTable := new(eai.RateTable)
	err = app.System(sv.LockedRateTableName, lockedTable)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("Error fetching %s system variable in CreditEAI.Apply", sv.LockedRateTableName))
	}

	feeTable := new(sv.EAIFeeTable)
	err = app.System(sv.EAIFeeTableName, feeTable)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("Error fetching %s system variable in CreditEAI.Apply", sv.EAIFeeTableName))
	}

	// calculate the actual award per ndau of EAI, so we can reduce each account's
	// award appropriately
	awardPerNdau := math.Ndau(constants.QuantaPerUnit)
	for _, fee := range *feeTable {
		awardPerNdau -= fee.Fee
	}

	// accumulate the total EAI credited by this transaction so we can award
	// fees appropriately
	var totalEAICredited uint64

	return app.UpdateState(func(stateI metast.State) (metast.State, error) {
		state := stateI.(*backing.State)
		nodeData, _ := state.GetAccount(tx.Node, app.blockTime)

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

			eaiAward, err = eaiAward.Add(acctData.UncreditedEAI)
			if err != nil {
				errorList = append(errorList, err)
				err = nil
				continue
			}

			// add the total EAI credited BEFORE reducing it
			totalEAICredited += uint64(eaiAward)

			// now reduce the award to account for the fees
			reducedAward, err := signed.MulDiv(
				int64(eaiAward),
				int64(awardPerNdau),
				constants.QuantaPerUnit,
			)
			if err != nil {
				errorList = append(errorList, err)
				err = nil
				continue
			}
			eaiAward = math.Ndau(reducedAward)
			_, err = state.PayReward(accountAddr, eaiAward, app.blockTime, true)
			if err != nil {
				errorList = append(errorList, err)
				err = nil
				continue
			}

		}

		// before considering the error list generated from the account iteration,
		// we want to ensure that appropriate fees get credited regardless
		for _, fee := range *feeTable {
			feeAward, err := signed.MulDiv(
				int64(totalEAICredited),
				int64(fee.Fee),
				constants.QuantaPerUnit,
			)
			if err != nil {
				errorList = append(errorList, err)
				err = nil
				continue
			}
			if fee.To == nil {
				state.PendingNodeReward, err = state.PendingNodeReward.Add(math.Ndau(feeAward))
				if err != nil {
					errorList = append(errorList, errors.Wrap(
						err,
						"adding unclaimed node rewards",
					))
					err = nil
					continue
				}
			} else {
				feeAcct, _ := state.GetAccount(*fee.To, app.blockTime)
				feeAcct.Balance, err = feeAcct.Balance.Add(math.Ndau(feeAward))
				if err != nil {
					errorList = append(errorList, err)
					err = nil
					continue
				}
				// Because this is a required state update, once we get this far,
				// we MUST NOT return an error from this function.
				state.Accounts[fee.To.String()] = feeAcct
			}
		}

		// Since the comments above indicate that the desire is to make sure the state gets
		// propagated even though there are errors, I'm going to suppress the error return
		// here since the caller will not update state if error is non-nil.
		// (See app.UpdateState in metanode/pkg/meta/app/application.go)
		if len(errorList) > 0 {
			errStr := fmt.Sprintf("Errors found calculating EAI for node %s: ", tx.Node.String())
			for idx, err := range errorList {
				if idx != 0 {
					errStr += ", "
				}
				errStr += err.Error()
			}
			err = errors.New(errStr)
			app.DecoratedTxLogger(tx).WithError(err).Error("CreditEAI.Apply() found errors; suppressing")
		}

		return state, nil
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

// GetSignatures implements signeder
func (tx *CreditEAI) GetSignatures() []signature.Signature {
	return tx.Signatures
}

// AppendSignatures implements Signable
func (tx *CreditEAI) AppendSignatures(sa []signature.Signature) {
	tx.Signatures = append(tx.Signatures, sa...)
}
