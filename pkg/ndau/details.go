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
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strings"

	metast "github.com/ndau/metanode/pkg/meta/state"
	metatx "github.com/ndau/metanode/pkg/meta/transaction"
	"github.com/ndau/ndau/pkg/ndau/backing"
	"github.com/ndau/ndaumath/pkg/address"
	"github.com/ndau/ndaumath/pkg/bitset256"
	"github.com/ndau/ndaumath/pkg/eai"
	"github.com/ndau/ndaumath/pkg/signature"
	math "github.com/ndau/ndaumath/pkg/types"
	sv "github.com/ndau/system_vars/pkg/system_vars"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

// MaxSequenceIncrement is the max allowed difference between the current
// sequence number and its successor
const MaxSequenceIncrement = 128

// A Sourcer is a transaction with a source from which tx fees are withdrawn,
// whose sequence number is checked, etc.
type Sourcer interface {
	// return the address which is considered the transaction's source
	// this address is used to check the sequence number, provides the tx
	// fees, etc.
	GetSource(*App) (address.Address, error)
}

// A Sequencer is a transaction with a sequence number.
type Sequencer interface {
	// return the sequence number of this transaction
	GetSequence() uint64
}

// A Withdrawer is a transaction which withdraws some amount from the source
// in addition to the normal tx fees.
type Withdrawer interface {
	// return the amount withdrawn from the source by this transaction.
	// Does not include tx fee or SIB.
	Withdrawal() math.Ndau
}

// A Signeder is a transaction which has some number of signatures.
type Signeder interface {
	// return the signatures of this transaction
	GetSignatures() []signature.Signature
}

// Signable allows signatures to be added to this TX
type Signable interface {
	// append to the list of signatures on this transaction
	ExtendSignatures([]signature.Signature)
}

// NTransactable is a wrapper around metatx.Transactable that allows us to
// manipulate ndau tx specifically. It's a little confusing to just call it
// Transactable.
type NTransactable interface {
	metatx.Transactable
	Sourcer
	Sequencer
}

// getTxAccount gets and validates an account for a transactable
//
// It returns a nil error if all of:
//  - sequence number is high enough
//  - account contains enough available ndau to pay the transaction fee
//  - if validation script set:
//       validation script passes
//  - if tx implements Signeder:
//       1 of N signature validation passes
//  - if tx implements Withdrawer:
//       account contains enough available ndau to pay the withdrawal + tx fee
func (app *App) getTxAccount(tx NTransactable) (backing.AccountData, bool, *bitset256.Bitset256, error) {
	validateScript := func(acct backing.AccountData, sigset *bitset256.Bitset256) error {
		if len(acct.ValidationScript) > 0 {
			vm, err := BuildVMForTxValidation(acct.ValidationScript, acct, tx, sigset, app)
			if err != nil {
				return errors.Wrap(err, "couldn't build vm for validation script")
			}
			err = vm.Run(nil)
			if err != nil {
				return errors.Wrap(err, "validation script")
			}
			vmReturn, err := vm.Stack().PopAsInt64()
			if err != nil {
				return errors.Wrap(err, "validation script exited without numeric stack top")
			}
			if vmReturn != 0 {
				return errors.New("validation script exited with non-0 exit code")
			}
		}
		return nil
	}

	addr, err := tx.GetSource(app)
	if err != nil {
		return backing.AccountData{}, false, nil, err
	}

	acct, exists := app.getAccount(addr)

	// revalidate all addresses in the tx, because that isn't
	// handled by default by the msgpack deserializer
	{
		txv := reflect.ValueOf(tx)
		for txv.Kind() == reflect.Ptr || txv.Kind() == reflect.Interface {
			txv = txv.Elem()
		}
		if txv.Kind() == reflect.Struct {
			failures := make(map[string]error)
			addrtype := reflect.TypeOf(address.Address{})
			for fn := 0; fn < txv.NumField(); fn++ {
				fv := txv.Field(fn)
				if fv.Type() == addrtype {
					addr := fv.Interface().(address.Address)
					err := addr.Revalidate()
					if err != nil {
						failures[txv.Type().Field(fn).Name] = err
					}
				}
			}
			if len(failures) > 0 {
				fkeys := make([]string, 0, len(failures))
				for key := range failures {
					fkeys = append(fkeys, key)
				}
				sort.Strings(fkeys)

				ferrs := make([]string, 0, len(failures))
				for _, fkey := range fkeys {
					ferrs = append(ferrs, fmt.Sprintf("%s: %s", fkey, failures[fkey]))
				}

				return acct, exists, nil, errors.New("the following fields failed to validate: " + strings.Join(ferrs, ", "))
			}
		}
	}

	if tx.GetSequence() <= acct.Sequence {
		return acct, exists, nil, errors.New("sequence too low")
	}
	if app.IsFeatureActive("SequenceIncrementProtection") && tx.GetSequence() > acct.Sequence+MaxSequenceIncrement {
		return acct, exists, nil, errors.New("sequence too high")
	}

	var sigset *bitset256.Bitset256
	if signed, isSigned := tx.(Signeder); isSigned {
		var validates bool
		validates, sigset = acct.ValidateSignatures(
			tx.SignableBytes(),
			signed.GetSignatures(),
		)
		if !validates {
			return acct, exists, sigset, fmt.Errorf("invalid signature(s): %d", sigset.Indices())
		}
	}

	err = validateScript(acct, sigset)
	if err != nil {
		return acct, exists, sigset, err
	}

	fee, err := app.calculateTxFee(tx)
	if err != nil {
		return acct, exists, sigset, err
	}
	if w, isWithdrawer := tx.(Withdrawer); isWithdrawer {
		fee, err = fee.Add(w.Withdrawal())
		if err != nil {
			return acct, exists, sigset, err
		}
	}

	sib, err := app.calculateSIB(tx)
	if err != nil {
		return acct, exists, sigset, err
	}
	fee, err = fee.Add(sib)
	if err != nil {
		return acct, exists, sigset, err
	}

	acct.UpdateRecourses(app.BlockTime())
	available, err := acct.AvailableBalance()
	if err != nil {
		return acct, exists, sigset, err
	}

	if available.Compare(fee) < 0 {
		var errb []byte
		errb, err = json.Marshal(map[string]interface{}{
			"balance":           acct.Balance,
			"available balance": available,
			"address":           addr,
			"tx cost":           fee,
			"msg":               "insufficient available balance to pay for tx",
			"tx hash":           metatx.Hash(tx),
			"qty holds":         len(acct.Holds),
		})
		if err == nil {
			err = errors.New(string(errb))
		} else {
			// we don't really care what the marshalling error was, and it's
			// unlikely ever to occur, but we left the old code in here just in case
			err = fmt.Errorf(
				"%s: insufficient available balance (%s ndau) to pay for tx (%s ndau)",
				addr,
				available,
				fee,
			)
		}

		return acct, exists, sigset, err
	}

	return acct, exists, sigset, err
}

// Every transaction has a transaction fee. This implies that every transaction
// touches the account balance somehow. This in turn implies that for our EAI
// calculations to work properly, we need to update them for every transaction.
//
// The details which get handled:
//  - update uncredited EAI with current balance
//  - deduct tx fee
//  - reduce source balance (if applicable)
//  - if source drops below 1000 ndau, remove currency seat
//  - update sequence
//  - resolve completed recourse holds
//
// Of course, most transactions will imply more modifications than these, but
// this at least provides a standard template for taking care of the basics.
//
// This function returns a closure encapsulating these updates. It is such that
// if a tx has no other effects on state, it can be embedded directly into a
// UpdateState function call. Otherwise, it can be run within such a call's
// embedded closure.
//
// The motivation for this design is to prevent a half-applied tx. If a tx's
// Apply method calls UpdateState more than once, then there exists the
// potential for a partially applied transaction, if it errors out in between
// the first call and the second. Therefore, it is important that each
// transaction's Apply method calls UpdateState either 0 or 1 times. Therefore,
// in order to apply the work we desire in this tx, we must return a closure
// instead of directly updating the state ourselves.
//
// This function assumes that all necessary validation (such as occurs in
// getTxAccount) has already been performed.
func (app *App) applyTxDetails(tx NTransactable) func(metast.State) (metast.State, error) {
	return func(stI metast.State) (metast.State, error) {
		if tx == nil {
			return stI, errors.New("nil transactable")
		}

		fee, err := app.calculateTxFee(tx)
		if err != nil {
			return stI, errors.Wrap(err, "calculating tx fee")
		}

		sib, err := app.calculateSIB(tx)
		if err != nil {
			return stI, errors.Wrap(err, "calculating SIB")
		}

		sourceA, err := tx.GetSource(app)
		if err != nil {
			return stI, errors.Wrap(err, "getting tx source")
		}
		sourceS := sourceA.String()

		unlockedTable := new(eai.RateTable)
		err = app.System(sv.UnlockedRateTableName, unlockedTable)
		if err != nil {
			return stI, errors.Wrap(err, fmt.Sprintf("Error fetching %s system variable in applyTxDetails", sv.UnlockedRateTableName))
		}
		lockedTable := new(eai.RateTable)
		err = app.System(sv.LockedRateTableName, lockedTable)
		if err != nil {
			return stI, errors.Wrap(err, fmt.Sprintf("Error fetching %s system variable in applyTxDetails", sv.UnlockedRateTableName))
		}

		source, _ := app.getAccount(sourceA)

		// if source isn't locked, resets the lock data
		source.IsLocked(app.BlockTime())

		pending, err := source.Balance.Add(source.UncreditedEAI)
		if err != nil {
			return stI, errors.Wrap(err, "adding uncredited eai to balance for new eai calc")
		}

		if app.IsFeatureActive("ApplyUncreditedEAI") {
			err = source.WeightedAverageAge.UpdateWeightedAverageAge(
				app.BlockTime().Since(source.LastWAAUpdate),
				0,
				source.Balance,
			)
			if err != nil {
				return stI, errors.Wrap(err, "updating weighted average age")
			}
		}

		logger := app.GetLogger().WithFields(log.Fields{
			"sourceAcct":         sourceA.String(),
			"pending":            pending.String(),
			"blockTime":          app.BlockTime().String(),
			"lastEAIUpdate":      source.LastEAIUpdate.String(),
			"weightedAverageAge": source.WeightedAverageAge.String(),
		})
		if source.Lock == nil {
			logger = logger.WithField("lock", "nil")
		} else {
			logger = logger.WithFields(log.Fields{
				"lock.noticePeriod": source.Lock.NoticePeriod.String(),
				"lock.bonus":        source.Lock.Bonus.String(),
			})
			if source.Lock.UnlocksOn == nil {
				logger = logger.WithField("lock.unlocks on", "nil")
			} else {
				logger = logger.WithField("lock.unlocks on", source.Lock.UnlocksOn.String())
			}
		}
		logger.Info("details eai calculation fields")

		eai, err := eai.Calculate(
			pending, app.BlockTime(), source.LastEAIUpdate,
			source.WeightedAverageAge, source.Lock,
			*unlockedTable, app.IsFeatureActive("FixEAIUnlockBug"),
		)
		if err != nil {
			return stI, errors.Wrap(err, "calculating uncredited eai")
		}

		source.UncreditedEAI, err = source.UncreditedEAI.Add(eai)
		if err != nil {
			return stI, errors.Wrap(err, "summing uncredited eai")
		}
		source.LastEAIUpdate = app.BlockTime()

		withdrawal, err := fee.Add(sib)
		if err != nil {
			return stI, errors.Wrap(err, "adding fee and sib")
		}
		if w, isWithdrawer := tx.(Withdrawer); isWithdrawer {
			withdrawal, err = withdrawal.Add(w.Withdrawal())
			if err != nil {
				return stI, errors.Wrap(err, "adding withdrawal qty to fees")
			}
		}

		source.Balance, err = source.Balance.Sub(withdrawal)
		if err != nil {
			return stI, errors.Wrap(err, "calculating new balance")
		}

		source.UpdateCurrencySeat(app.BlockTime())

		source.Sequence = tx.GetSequence()

		source.UpdateRecourses(app.BlockTime())

		st := stI.(*backing.State)
		st.Accounts[sourceS] = source
		st.PendingNodeReward += fee
		st.TotalBurned += sib
		return st, nil
	}
}

// AddressIndexable is a Transactable that has addresses associated with it that we want to index.
type AddressIndexable interface {
	metatx.Transactable
	GetAccountAddresses(*App) ([]string, error)
}

// GetAccountAddresses gets the affected account addresses from a tx
//
// Transactions can override the behavior by implementing AddressIndexable,
// but by default, every ndau transactable will return its source
func (app *App) GetAccountAddresses(tx metatx.Transactable) ([]string, error) {
	switch x := tx.(type) {
	case AddressIndexable:
		return x.GetAccountAddresses(app)
	case Sourcer:
		addr, err := x.GetSource(app)
		return []string{addr.String()}, err
	default:
		// if we only ever use NTransactables, this will never happen
		return []string{}, nil
	}
}
