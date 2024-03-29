package ndau

// ----- ---- --- -- -
// Copyright 2019 Oneiro NA, Inc. All Rights Reserved.
//
// Licensed under the Apache License 2.0 (the "License").  You may not use
// this file except in compliance with the License.  You can obtain a copy
// in the file LICENSE in the source distribution or at
// https://www.apache.org/licenses/LICENSE-2.0.txt
// - -- --- ---- -----

import (
	"fmt"
	"sort"
	"strings"

	metast "github.com/ndau/metanode/pkg/meta/state"
	"github.com/ndau/ndau/pkg/ndau/backing"
	srch "github.com/ndau/ndau/pkg/ndau/search"
	"github.com/ndau/ndaumath/pkg/address"
	"github.com/ndau/ndaumath/pkg/constants"
	"github.com/ndau/ndaumath/pkg/eai"
	"github.com/ndau/ndaumath/pkg/signature"
	"github.com/ndau/ndaumath/pkg/signed"
	math "github.com/ndau/ndaumath/pkg/types"
	sv "github.com/ndau/system_vars/pkg/system_vars"
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

	/*
		2022-05-01 Change to use current lock bonus table, not the bonus value stored in the account. Set a flag
		here to be used later, because we have to do this test for every account in the credit EAI loop. The rate
		table could change at any time, so we always have to check.
	*/

	lockedBonusRateTable := eai.RateTable{}
	useCurrentLockBonus := app.IsFeatureActive("UseCurrentLockBonus")
	if useCurrentLockBonus {
		err = app.System(sv.LockedRateTableName, &lockedBonusRateTable)
		if err != nil {
			return err
		}
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

	return app.UpdateState(
		app.recalculateWAAs(tx),
		app.applyTxDetails(tx),
		func(stateI metast.State) (metast.State, error) {
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

					/*
						2022-05-01 Change - always use the current lock bonus, not the saved one, and
						update the account state with the new rate if it's different.
					*/

					if useCurrentLockBonus {
						currentLockBonus := lockedBonusRateTable.RateAt(acctData.Lock.NoticePeriod)
						/*
							2023-05-02 Remove uninteresting logging message for EVERY locked account on EVERY CreditEAI transaction!

									app.DecoratedTxLogger(tx).WithFields(log.Fields{
									"acctLockBonus":    acctData.Lock.Bonus.String(),
									"currentLockBonus": currentLockBonus.String(),
								}).Info("lock bonus update")
						*/
						if acctData.Lock.Bonus != currentLockBonus {
							acctData.Lock.Bonus = currentLockBonus
							state.Accounts[addrS].Lock.Bonus = acctData.Lock.Bonus
						}
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

			// JSG the above might have modified total ndau in circulation, so recalculate SIB
			if app.IsFeatureActive("AllRFEInCirculation") {
				sib, target, err := app.calculateCurrentSIB(state, -1, -1)
				if err != nil {
					return state, err
				}
				state.SIB = sib
				state.TargetPrice = target
			}

			// Since the comments above indicate that the desire is to make sure the state gets
			// propagated even though there are errors, I'm going to suppress the error return
			// here since the caller will not update state if error is non-nil.
			// (See app.UpdateState in metanode/pkg/meta/app/application.go)

			return state, nil
		},
	)
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

func (app *App) recalc(stateI metast.State, feature string, tx *CreditEAI) (metast.State, error) {
	state := stateI.(*backing.State)
	indexer := app.GetSearch()
	if app.IsFeatureActive(feature) && indexer != nil {
		// the feature height may or may not be set. Our goal is to run these
		// calculations exactly once, on the first CreditEAI after the feature
		// gate is set. The methodology is pretty simple:
		featureHeight := uint64(0)
		if app.config.Features != nil {
			featureHeight = app.config.Features[feature]
		}

		client := indexer.(*srch.Client)
		featureTs, err := client.BlockTime(featureHeight)
		for featureTs == math.Timestamp(0) && err == nil {
			featureHeight++
			if featureHeight == app.Height() {
				featureTs = app.BlockTime()
			} else {
				featureTs, err = client.BlockTime(featureHeight)
			}
		}
		if err != nil {
			// if we can't recalculate the WAAs, we return an error here, which
			// prevents any further calculation of EAI credits. We get this right,
			// or the transaction doesn't succeed at all.
			return stateI, fmt.Errorf(
				"failed to get block time for height %d: %w",
				featureHeight,
				err,
			)
		}

		logger := app.DecoratedLogger()
		logger.WithFields(log.Fields{
			"featureTs":     featureTs,
			"featureHeight": featureHeight,
		}).Info("WAA update begin")

		// from here, it's straightforward: for every account, if its WAA was
		// last calculated before the feature timestamp, recalculate its WAA
		// from genesis. It would be better to avoid iteration if possible,
		// but there are only a few thousand accounts; the cost of iterating
		// over them again and doing nothing is likely to be dwarfed by the
		// rest of the CreditEAI calculation

		delegatedAccounts := state.Delegates[tx.Node.String()]

		// If there are no accounts delegated to the node, there's nothing to do
		if len(delegatedAccounts) == 0 {
			return state, nil
		}

		delegatedAccounts[tx.Node.String()] = struct{}{}

		for addrS := range delegatedAccounts {
			addr, err := address.Validate(addrS)
			if err != nil {
				// if we can't recalculate the WAAs, we return an error here, which
				// prevents any further calculation of EAI credits. We get this right,
				// or the transaction doesn't succeed at all.
				return stateI, fmt.Errorf(
					"failed to get block time for height %d: %w",
					featureHeight,
					err,
				)
			}

			acctData, hasAcct := app.getAccount(addr)
			if !hasAcct {
				// Accounts might sometimes be removed.
				// If we encounter that, don't worry about it. An account
				// which doesn't exist necessarily has 0 balance and is
				// therefore ineligible to receive EAI anyway. Likewise,
				// it can't have an incoming rewards list of len > 0.
				// We can therefore return early here as a minor optimization.
				continue
			}
			// logger.WithFields(log.Fields{
			// 	"addr":          addrS,
			// 	"LastWAAUpdate": acctData.LastWAAUpdate,
			// 	"WAA":           acctData.WeightedAverageAge,
			// }).Info("WAA update addr")

			if acctData.LastWAAUpdate < featureTs {
				ahr, err := client.SearchAccountHistory(addrS, 0, 0)
				if err != nil {
					return stateI, fmt.Errorf("failed to query tx history: %w", err)
				}

				// logger.WithFields(log.Fields{
				// 	"addr": addrS,
				// 	"ahr":  ahr,
				// }).Info("WAA update ahr")

				// this lastWAAUpdate, unlike the one in the account, should
				// be correct at all times
				var waa math.Duration
				var lastWAAUpdate math.Timestamp
				var lastBalance math.Ndau
				// the following keeps track if we've seen a TX that sets a balance, this helps us to determine if
				// this acct is a Transfer destination
				balanceSeen := false

				for _, valueData := range ahr.Txs {
					balance := valueData.Balance
					blockTime, err := client.BlockTime(valueData.BlockHeight)
					// err = rows.Scan(&blockTime, &name, &hash, &data, &balance)
					if err != nil {
						return stateI, fmt.Errorf(
							"failed to get BlockTime for %s: %w",
							addrS,
							err,
						)
					}

					// account creation gets a bit of special handling
					if lastWAAUpdate == 0 {
						lastWAAUpdate = blockTime
					}

					// all transactions have the Details waa update applied first
					// this makes uncredited EAI work better; I don't remember
					// exactly why and the comment isn't explicit, but we do it
					// sincePrev := math.DurationFrom(blockTime.Sub(lastWAAUpdate))
					sincePrev := blockTime.Since(lastWAAUpdate)
					waa.UpdateWeightedAverageAge(
						sincePrev,
						0,
						math.Ndau(balance),
					)

					// the WAA is also handled specially in Transfer TXs
					txData, err := client.GetTxTypeAtHeight(valueData.BlockHeight, "Transfer", 0)
					if err != nil {
						return stateI, fmt.Errorf(
							"failed to get txData for Transfer at block height %d for %s: %w",
							valueData.BlockHeight,
							addrS,
							err,
						)
					}

					// if we found a Transfer TX, deal with it appropriately
					if len(txData.Txs) > 0 {
						// if we've seen the balance set, and we are the destination, set the WAA
						if balanceSeen && balance > lastBalance {
							waa.UpdateWeightedAverageAge(
								sincePrev,
								balance-lastBalance,
								lastBalance,
							)
						}
					}

					// update loop variables
					lastBalance = balance
					balanceSeen = true
					lastWAAUpdate = blockTime
				}

				logger.WithFields(log.Fields{
					"addr":          addrS,
					"WAA":           waa,
					"LastWAAUpdate": app.BlockTime(),
				}).Info("WAA update addr")

				// we've recalculated the full history for this account; now
				// just plug the values back in
				acctData.WeightedAverageAge = waa
				if app.IsFeatureActive("FixLastWAAUpdate") {
					acctData.LastWAAUpdate = lastWAAUpdate
				} else {
					acctData.LastWAAUpdate = app.BlockTime()
				}
				state.Accounts[addrS] = acctData
			}
		}
		logger.Info("WAA update end")
	}
	return state, nil
}

// we had a WAA bug which caused some nonsense weighted average ages
// this function runs on each CreditEAI to ensure that we have sensible values
func (app *App) recalculateWAAs(tx *CreditEAI) func(metast.State) (metast.State, error) {
	return func(stateI metast.State) (metast.State, error) {
		var err error
		state := stateI
		// run this for original WAA fix
		feature := "UpdateWAAUpdateDateInDetails"
		if app.IsFeatureActive(feature) {
			state, err = app.recalc(state, feature, tx)
		}
		// run this to fix LastWAAUpdate value
		feature = "FixLastWAAUpdate"
		if app.IsFeatureActive(feature) {
			state, err = app.recalc(state, feature, tx)
		}
		return state, err
	}
}
