package ndau

import (
	"encoding/json"
	"fmt"

	metast "github.com/oneiro-ndev/metanode/pkg/meta/state"
	metatx "github.com/oneiro-ndev/metanode/pkg/meta/transaction"
	"github.com/oneiro-ndev/ndau/pkg/ndau/backing"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	"github.com/oneiro-ndev/ndaumath/pkg/bitset256"
	"github.com/oneiro-ndev/ndaumath/pkg/eai"
	"github.com/oneiro-ndev/ndaumath/pkg/signature"
	math "github.com/oneiro-ndev/ndaumath/pkg/types"
	sv "github.com/oneiro-ndev/system_vars/pkg/system_vars"
	"github.com/pkg/errors"
)

type sourcer interface {
	// return the address which is considered the transaction's source
	// this address is used to check the sequence number, provides the tx
	// fees, etc.
	GetSource(*App) (address.Address, error)
}

type sequencer interface {
	// return the sequence number of this transaction
	GetSequence() uint64
}

type withdrawer interface {
	// return the amount withdrawn from the source by this transaction.
	// Does not include tx fee or SIB.
	Withdrawal() math.Ndau
}

type signeder interface {
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
	sourcer
	sequencer
}

// getTxAccount gets and validates an account for a transactable
//
// It returns a nil error if all of:
//  - sequence number is high enough
//  - account contains enough available ndau to pay the transaction fee
//  - if validation script set:
//       validation script passes
//  - if tx implements signeder:
//       1 of N signature validation passes
//  - if tx implements withdrawer:
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

	address, err := tx.GetSource(app)
	if err != nil {
		return backing.AccountData{}, false, nil, err
	}

	acct, exists := app.getAccount(address)
	if tx.GetSequence() <= acct.Sequence {
		return acct, exists, nil, errors.New("sequence too low")
	}

	var sigset *bitset256.Bitset256
	if signed, isSigned := tx.(signeder); isSigned {
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
	if w, isWithdrawer := tx.(withdrawer); isWithdrawer {
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

	acct.UpdateSettlements(app.blockTime)
	available, err := acct.AvailableBalance()
	if err != nil {
		return acct, exists, sigset, err
	}

	if available.Compare(fee) < 0 {
		var errb []byte
		errb, err = json.Marshal(map[string]interface{}{
			"balance":                 acct.Balance,
			"available balance":       available,
			"address":                 address,
			"tx cost":                 fee,
			"msg":                     "insufficient available balance to pay for tx",
			"tx hash":                 metatx.Hash(tx),
			"qty pending settlements": len(acct.Settlements),
		})
		if err == nil {
			err = errors.New(string(errb))
		} else {
			// we don't really care what the marshalling error was, and it's
			// unlikely ever to occur, but we left the old code in here just in case
			err = fmt.Errorf(
				"%s: insufficient available balance (%s ndau) to pay for tx (%s ndau)",
				address,
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
//  - resolve completed settlements
//
// Of course, most transactions will imply more modifications than these, but
// this at least provides a standard template for taking care of the basics.
//
// If the return value is not nil, this function guarantees that it will not
// have modified the app state.
//
// This function should only be called in Apply implementations; it assumes
// that all necessary validation (such as occurs in getTxAccount) has already
// been performed.
func (app *App) applyTxDetails(tx NTransactable) error {
	if tx == nil {
		return errors.New("nil transactable")
	}

	fee, err := app.calculateTxFee(tx)
	if err != nil {
		return errors.Wrap(err, "calculating tx fee")
	}

	sib, err := app.calculateSIB(tx)
	if err != nil {
		return errors.Wrap(err, "calculating SIB")
	}

	sourceA, err := tx.GetSource(app)
	if err != nil {
		return errors.Wrap(err, "getting tx source")
	}
	sourceS := sourceA.String()

	unlockedTable := new(eai.RateTable)
	err = app.System(sv.UnlockedRateTableName, unlockedTable)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("Error fetching %s system variable in applyTxDetails", sv.UnlockedRateTableName))
	}
	lockedTable := new(eai.RateTable)
	err = app.System(sv.LockedRateTableName, lockedTable)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("Error fetching %s system variable in applyTxDetails", sv.UnlockedRateTableName))
	}

	source, _ := app.getAccount(sourceA)

	eai, err := eai.Calculate(
		source.Balance, app.blockTime, source.LastEAIUpdate,
		source.WeightedAverageAge, source.Lock,
		*unlockedTable,
	)

	source.UncreditedEAI, err = source.UncreditedEAI.Add(eai)
	if err != nil {
		return errors.Wrap(err, "calculating new uncredited EAI")
	}
	source.LastEAIUpdate = app.blockTime

	withdrawal, err := fee.Add(sib)
	if err != nil {
		return errors.Wrap(err, "adding fee and sib")
	}
	if w, isWithdrawer := tx.(withdrawer); isWithdrawer {
		withdrawal, err = withdrawal.Add(w.Withdrawal())
		if err != nil {
			return errors.Wrap(err, "adding withdrawal qty to fees")
		}
	}

	source.Balance, err = source.Balance.Sub(withdrawal)
	if err != nil {
		return errors.Wrap(err, "calculating new balance")
	}

	source.UpdateCurrencySeat(app.blockTime)

	source.Sequence = tx.GetSequence()

	/////////////////////////////////////////////////////////////////////
	// Everything which may return an error must go above this line.   //
	// Below this point, no error values are permitted.                //
	/////////////////////////////////////////////////////////////////////

	source.UpdateSettlements(app.blockTime)

	return app.UpdateState(func(stI metast.State) (metast.State, error) {
		st := stI.(*backing.State)
		st.Accounts[sourceS] = source
		st.TotalFees += fee
		st.TotalSIB += sib
		return st, nil
	})
}
