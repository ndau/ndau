package ndau

import (
	"fmt"

	metast "github.com/oneiro-ndev/metanode/pkg/meta/state"
	metatx "github.com/oneiro-ndev/metanode/pkg/meta/transaction"
	"github.com/oneiro-ndev/ndau/pkg/ndau/backing"
	sv "github.com/oneiro-ndev/ndau/pkg/ndau/system_vars"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	"github.com/oneiro-ndev/ndaumath/pkg/eai"
	math "github.com/oneiro-ndev/ndaumath/pkg/types"
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

type ndauTransactable interface {
	metatx.Transactable
	sourcer
	sequencer
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
		*unlockedTable, *lockedTable,
	)

	source.UncreditedEAI, err = source.UncreditedEAI.Add(eai)
	if err != nil {
		return errors.Wrap(err, "calculating new uncredited EAI")
	}

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

	return app.UpdateState(func(stI metast.State) (metast.State, error) {
		st := stI.(*backing.State)
		st.Accounts[sourceS] = source
		return st, nil
	})
}
