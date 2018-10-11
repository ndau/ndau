package ndau

import (
	"fmt"

	metast "github.com/oneiro-ndev/metanode/pkg/meta/state"
	metatx "github.com/oneiro-ndev/metanode/pkg/meta/transaction"
	"github.com/oneiro-ndev/ndau/pkg/ndau/backing"
	sv "github.com/oneiro-ndev/ndau/pkg/ndau/system_vars"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	"github.com/oneiro-ndev/ndaumath/pkg/bitset256"
	"github.com/oneiro-ndev/ndaumath/pkg/eai"
	math "github.com/oneiro-ndev/ndaumath/pkg/types"
	"github.com/oneiro-ndev/ndaumath/pkg/signature"
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

type ndauTransactable interface {
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
func (app *App) getTxAccount(tx ndauTransactable) (backing.AccountData, bool, *bitset256.Bitset256, error) {
	validateScript := func(acct backing.AccountData, sigset *bitset256.Bitset256) error {
		if len(acct.ValidationScript) > 0 {
			vm, err := BuildVMForTxValidation(acct.ValidationScript, acct, tx, sigset, app)
			if err != nil {
				return errors.Wrap(err, "couldn't build vm for validation script")
			}
			err = vm.Run(false)
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

	state := app.GetState().(*backing.State)

	acct, exists := state.GetAccount(address, app.blockTime)
	if tx.GetSequence() <= acct.Sequence {
		return acct, exists, nil, errors.New("Sequence too low")
	}

	var sigset *bitset256.Bitset256
	if signed, isSigned := tx.(signeder); isSigned {
		var validates bool
		validates, sigset = acct.ValidateSignatures(
			tx.SignableBytes(),
			signed.GetSignatures(),
		)
		if !validates {
			return acct, exists, sigset, errors.New("Invalid signature(s)")
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

	acct.UpdateSettlements(app.blockTime)
	available, err := acct.AvailableBalance()
	if err != nil {
		return acct, exists, sigset, err
	}

	if available.Compare(fee) < 0 {
		err = fmt.Errorf(
			"insufficient available balance (%s ndau) to pay for tx (%s ndau)",
			available,
			fee,
		)
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
func (app *App) applyTxDetails(tx ndauTransactable) error {
	if tx == nil {
		return errors.New("nil transactable")
	}

	fee, err := app.calculateTxFee(tx)
	if err != nil {
		return errors.Wrap(err, "calculating tx fee")
	}

	sourceA, err := tx.GetSource(app)
	if err != nil {
		return errors.Wrap(err, "getting tx source")
	}
	sourceS := sourceA.String()

	unlockedTable := new(eai.RateTable)
	err = app.System(sv.UnlockedRateTableName, unlockedTable)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("Error fetching %s system variable in CreditEAI.Apply", sv.UnlockedRateTableName))
	}
	lockedTable := new(eai.RateTable)
	err = app.System(sv.LockedRateTableName, lockedTable)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("Error fetching %s system variable in CreditEAI.Apply", sv.UnlockedRateTableName))
	}

	state := app.GetState().(*backing.State)
	source := state.Accounts[sourceS]

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

	withdrawal := fee
	if w, isWithdrawer := tx.(withdrawer); isWithdrawer {
		withdrawal, err = withdrawal.Add(w.Withdrawal())
		if err != nil {
			return errors.Wrap(err, "adding fee and withdrawal")
		}
	}

	source.Balance, err = source.Balance.Sub(withdrawal)
	if err != nil {
		return errors.Wrap(err, "calculating new balance")
	}

	source.Sequence = tx.GetSequence()

	/////////////////////////////////////////////////////////////////////
	// Everything which may return an error must go above this line.   //
	// Below this point, no error values are permitted.                //
	/////////////////////////////////////////////////////////////////////

	source.UpdateSettlements(app.blockTime)

	return app.UpdateState(func(stI metast.State) (metast.State, error) {
		st := stI.(*backing.State)
		st.Accounts[sourceS] = source
		return st, nil
	})
}
