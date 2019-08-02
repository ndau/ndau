package ndau

import (
	"fmt"
	"sort"
	"strings"

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

// Validate implements metatx.Transactable
func (tx *CreditEAI) Validate(appI interface{}) error {
	app := appI.(*App)

	_, hasNode, _, err := app.getTxAccount(tx)
	if err != nil {
		return err
	}

	if !hasNode {
		return errors.New("no such node")
	}

	// EAI node must be active
	if app.IsFeatureActive("NodeActiveCheck") && !app.GetState().(*backing.State).IsActiveNode(tx.Node) {
		return errors.New("node must be active and is not")
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

	eaiOvertime := new(math.Duration)
	err = app.System(sv.EAIOvertime, eaiOvertime)
	if err != nil {
		app.DecoratedTxLogger(tx).Info("could not get EAI Overtime sysvar; will not apply overtime limit")
		err = nil
		eaiOvertime = nil
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

	fixEAIUnlockBug := app.IsFeatureActive("FixEAIUnlockBug")

	return app.UpdateState(app.applyTxDetails(tx), func(stateI metast.State) (metast.State, error) {
		state := stateI.(*backing.State)
		nodeData, _ := app.getAccount(tx.Node)

		state.Accounts[tx.Node.String()] = nodeData
		delegatedAccounts := state.Delegates[tx.Node.String()]

		logger := app.DecoratedTxLogger(tx).WithFields(log.Fields{
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
				logger.WithError(err).Error("EAI.Calculate failed")
				err = nil
				return true
			}
			return false
		}

		calc := func(addrS string, postpone bool) {
			logger = logger.WithField("acct", addrS)
			addr, err := address.Validate(addrS)
			if handle(err) {
				return
			}

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
					"len_IncomingRewardsFrom",
					len(acctData.IncomingRewardsFrom),
				).Debug("postponing due to incoming rewards")
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

			// when the EAI overtime duration is set, this is the maximum amount
			// of EAI which can be applied by a CreditEAI transaction. This
			// encourages node operators to issue the tx regularly.
			lastUpdate := acctData.LastEAIUpdate
			if eaiOvertime != nil && lastUpdate.Add(*eaiOvertime) < app.BlockTime() {
				lastUpdate = app.BlockTime().Sub(*eaiOvertime)
			}

			tableRows := make([]string, 0, len(*ageTable))
			for _, row := range *ageTable {
				rt, _ := row.MarshalText()
				tableRows = append(tableRows, string(rt))
			}
			tableS := strings.Join(tableRows, "/")
			logger := app.DecoratedTxLogger(tx).WithFields(log.Fields{
				"sourceAcct":         addrS,
				"pending":            pending.String(),
				"lastUpdate":         lastUpdate.String(),
				"weightedAverageAge": acctData.WeightedAverageAge.String(),
				"ageTable":           tableS,
			})
			if acctData.Lock == nil {
				logger = logger.WithField("lock", "nil")
			} else {
				logger = logger.WithFields(log.Fields{
					"lock.noticePeriod": acctData.Lock.NoticePeriod.String(),
					"lock.bonus":        acctData.Lock.Bonus.String(),
				})
				if acctData.Lock.UnlocksOn == nil {
					logger = logger.WithField("lock.unlocksOn", "nil")
				} else {
					logger = logger.WithField("lock.unlocksOn", acctData.Lock.UnlocksOn.String())
				}
			}
			logger.Debug("credit EAI calculation fields")

			eaiAward, err := eai.Calculate(
				pending, app.BlockTime(), lastUpdate,
				acctData.WeightedAverageAge, acctData.Lock,
				*ageTable, fixEAIUnlockBug,
			)
			if handle(err) {
				return
			}

			app.DecoratedTxLogger(tx).WithFields(log.Fields{
				"sourceAcct":    addrS,
				"EAIAward":      eaiAward.String(),
				"uncreditedEAI": acctData.UncreditedEAI.String(),
			}).Debug("credit EAI calculation results")

			if app.IsFeatureActive("CreditEAIUnlocksAccounts") {
				// now that the lock data has been used to calculate the pending EAI,
				// clear it if it has expired.
				acctData.IsLocked(app.BlockTime())
				// we can't unconditionally update WAA, so we have to update only
				// the account data lock field
				ad2, ok := state.Accounts[addrS]
				if ok {
					ad2.Lock = acctData.Lock
					state.Accounts[addrS] = ad2
				}
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

			app.DecoratedTxLogger(tx).WithFields(log.Fields{
				"sourceAcct":   addrS,
				"totalAward":   eaiAward.String(),
				"reducedAward": math.Ndau(reducedAward).String(),
			}).Debug("credit EAI award reduction")

			eaiAward = math.Ndau(reducedAward)
			_, err = state.PayReward(
				addr,
				eaiAward,
				app.BlockTime(),
				app.getDefaultRecourseDuration(),
				true,
				app.IsFeatureActive("ResetUncreditedEAIOnCreditEAI"),
			)
			if handle(err) {
				return
			}
			logger.WithFields(log.Fields{
				"award":         eaiAward,
				"rewardsTarget": acctData.RewardsTarget,
			}).Debug("awarded EAI")
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
		logger.WithField("len_postponed", len(postponed)).Debug("calculating postponed accounts")
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

// GetAccountAddresses returns the account addresses associated with this transaction type.
func (tx *CreditEAI) GetAccountAddresses(app *App) ([]string, error) {
	state := app.GetState().(*backing.State)
	addrs := make([]string, 0, 1+len(state.Delegates[tx.Node.String()]))
	for d := range state.Delegates[tx.Node.String()] {
		addrs = append(addrs, d)
	}
	sort.Strings(addrs)
	addrs = append(addrs, tx.Node.String())

	return addrs, nil
}
