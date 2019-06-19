package ndau

import (
	"fmt"
	"sort"

	metast "github.com/oneiro-ndev/metanode/pkg/meta/state"
	"github.com/oneiro-ndev/ndau/pkg/ndau/backing"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	"github.com/oneiro-ndev/ndaumath/pkg/constants"
	"github.com/oneiro-ndev/ndaumath/pkg/eai"
	"github.com/oneiro-ndev/ndaumath/pkg/signature"
	"github.com/oneiro-ndev/ndaumath/pkg/signed"
	math "github.com/oneiro-ndev/ndaumath/pkg/types"
	sv "github.com/oneiro-ndev/system_vars/pkg/system_vars"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

// GetAccountAddresses returns the account addresses associated with this transaction type.
func (tx *CreditEAI) GetAccountAddresses() []string {
	return []string{tx.Node.String()}
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

	// Exchange accounts get a flat rate for EAI.  To accomplish this, we make a 1-element rate
	// table using the exchange account Rate (dependent on account) with a zero From field.
	exchangeTable := make(eai.RateTable, 1, 1)

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
		nodeData, _ := app.getAccount(tx.Node)

		state.Accounts[tx.Node.String()] = nodeData
		delegatedAccounts := state.Delegates[tx.Node.String()]

		logger := app.DecoratedTxLogger(tx).WithFields(log.Fields{
			"tx":            "CreditEAI",
			"node":          tx.Node.String(),
			"blockTime":     app.BlockTime(),
			"unlockedTable": unlockedTable,
		})

		// for deterministic EAI calculations, it is necessary that the
		// accounts receiving awards from elsewhere happen in a different
		// tranche than accounts not receiving awards from elsewhere.
		// It doesn't really matter logically which tranche comes first,
		// so we calculate the EAI of accounts receiving EAI from elsewhere
		// after the rest of them in order to maximize total EAI awarded.
		var postponed []string

		// we don't want to error out during CalculateEAI, because that could
		// result in not all accounts being considered, in a non-deterministic
		// fashion. Instead, log errors as they occur, but never fail here.
		handle := func(err error) bool {
			if err != nil {
				logger.WithError(err).Error("eai.Calculate failed")
				err = nil
				return true
			}
			return false
		}

		calc := func(addrS string, postpone bool) {
			addr, err := address.Validate(addrS)
			if handle(err) {
				return
			}
			logger = logger.WithField("acct", addr)

			acctData, hasAcct := app.getAccount(addr)
			if !hasAcct {
				// Accounts might sometimes be removed.
				// If we encounter that, don't worry about it. An account
				// which doesn't exist necessarily has 0 balance and is
				// therefore ineligible to receive EAI anyway. Likewise,
				// it can't have an incoming rewards list of len > 0.
				// We can therefore return early here as a minor optimization.
				return
			}

			if postpone && len(acctData.IncomingRewardsFrom) > 0 {
				logger.WithField(
					"len(IncomingRewardsFrom)",
					len(acctData.IncomingRewardsFrom),
				).Info("postponing due to incoming rewards")
				postponed = append(postponed, addrS)
				return
			}

			err = acctData.WeightedAverageAge.UpdateWeightedAverageAge(
				app.BlockTime().Since(acctData.LastWAAUpdate),
				0,
				acctData.Balance,
			)
			if handle(err) {
				return
			}
			acctData.LastWAAUpdate = app.BlockTime()

			// Select the appropriate age/rate table to use.
			ageTable := unlockedTable
			isExchangeAccount, err := app.GetState().(*backing.State).AccountHasAttribute(addr, sv.AccountAttributeExchange)
			if handle(err) {
				return
			}
			if isExchangeAccount {
				exchangeTable[0].Rate, err = app.calculateExchangeEAIRate(acctData)
				if handle(err) {
					return
				}
				ageTable = &exchangeTable
			}

			// we have to add the uncredited EAI to the balance before calculating
			// new EAI so that we grant the full amount. Failure to do so
			// means that people won't earn EAI on what is currently uncredited.
			pending, err := acctData.Balance.Add(acctData.UncreditedEAI)
			if handle(err) {
				return
			}

			eaiAward, err := eai.Calculate(
				pending, app.BlockTime(), acctData.LastEAIUpdate,
				acctData.WeightedAverageAge, acctData.Lock,
				*ageTable,
			)
			if handle(err) {
				return
			}

			eaiAward, err = eaiAward.Add(acctData.UncreditedEAI)
			if handle(err) {
				return
			}

			// add the total EAI credited BEFORE reducing it
			totalEAICredited += uint64(eaiAward)

			// now reduce the award to account for the fees
			reducedAward, err := signed.MulDiv(
				int64(eaiAward),
				int64(awardPerNdau),
				constants.QuantaPerUnit,
			)
			if handle(err) {
				return
			}
			eaiAward = math.Ndau(reducedAward)
			_, err = state.PayReward(
				addr,
				eaiAward,
				app.BlockTime(),
				app.getDefaultSettlementDuration(),
				true,
				app.IsFeatureActive(ResetUncreditedEAIOnCreditEAI),
			)
			if handle(err) {
				return
			}
			logger.WithFields(log.Fields{
				"award":         eaiAward,
				"rewardsTarget": acctData.RewardsTarget,
			}).Info("awarded EAI")
		}

		// for determinism, we must iterate the account list in a defined order
		// so we walk the map, record all the IDs, sort them, and then iterate that
		accountList := make([]string, 0, len(delegatedAccounts))
		for acct := range delegatedAccounts {
			accountList = append(accountList, acct)
		}
		sort.Sort(sort.StringSlice(accountList))

		// now iterate the account list deterministically
		for _, acct := range accountList {
			calc(acct, true)
		}

		// and finally do the postponed ones (in deterministic order as well)
		logger.WithField("len(postponed)", len(postponed)).Info("calculating postponed accounts")
		for _, acct := range postponed {
			calc(acct, false)
		}

		// before considering the error list generated from the account iteration,
		// we want to ensure that appropriate fees get credited regardless
		for _, fee := range *feeTable {
			feeAward, err := signed.MulDiv(
				int64(totalEAICredited),
				int64(fee.Fee),
				constants.QuantaPerUnit,
			)
			if handle(err) {
				continue
			}

			if fee.To == nil {
				state.PendingNodeReward, err = state.PendingNodeReward.Add(math.Ndau(feeAward))
				if handle(errors.Wrap(err, "adding unclaimed node rewards")) {
					continue
				}
			} else {
				feeAcct, _ := app.getAccount(*fee.To)
				feeAcct.Balance, err = feeAcct.Balance.Add(math.Ndau(feeAward))
				if handle(err) {
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

		return state, nil
	})
}

// GetSource implements Sourcer
func (tx *CreditEAI) GetSource(*App) (address.Address, error) {
	return tx.Node, nil
}

// GetSequence implements Sequencer
func (tx *CreditEAI) GetSequence() uint64 {
	return tx.Sequence
}

// GetSignatures implements Signeder
func (tx *CreditEAI) GetSignatures() []signature.Signature {
	return tx.Signatures
}

// ExtendSignatures implements Signable
func (tx *CreditEAI) ExtendSignatures(sa []signature.Signature) {
	tx.Signatures = append(tx.Signatures, sa...)
}
