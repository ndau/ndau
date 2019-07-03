package ndau

import (
	metast "github.com/oneiro-ndev/metanode/pkg/meta/state"
	metatx "github.com/oneiro-ndev/metanode/pkg/meta/transaction"
	"github.com/oneiro-ndev/ndau/pkg/ndau/backing"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	"github.com/oneiro-ndev/ndaumath/pkg/eai"
	"github.com/oneiro-ndev/ndaumath/pkg/signature"
	math "github.com/oneiro-ndev/ndaumath/pkg/types"
	sv "github.com/oneiro-ndev/system_vars/pkg/system_vars"
	"github.com/pkg/errors"
)

// GetAccountAddresses returns the account addresses associated with this transaction type.
func (tx *TransferAndLock) GetAccountAddresses() []string {
	return []string{tx.Source.String(), tx.Destination.String()}
}

// Validate satisfies metatx.Transactable
func (tx *TransferAndLock) Validate(appInt interface{}) error {
	app := appInt.(*App)

	if tx.Qty <= math.Ndau(0) {
		return errors.New("invalid transfer: Qty not positive")
	}

	if tx.Source == tx.Destination {
		return errors.New("invalid transfer: source == destination")
	}

	source, _, _, err := app.getTxAccount(tx)
	if err != nil {
		return err
	}

	if source.IsLocked(app.BlockTime()) {
		return errors.New("source is locked")
	}

	_, exists := app.getAccount(tx.Destination)

	// TransferAndLock cannot be sent to an existing account
	// This consequently also ensures that you cannot transfer and lock an exchange account.
	// The reason this is a sufficient check for that (and why we do not specifically check
	// whether the destination, if it's an exchange account, should not become locked), is
	// because being an exchange accounts implies that it exists.  An account cannot be an
	// exchange account and not exist, because an exchange accounts is defined to be one for
	// which its Progenitor is present in the AccountAttributes system variable and has the
	// AccountAttributeExchange flag associated with it there.  If getAccount() says that the
	// destination account doesn't exist, then it can't possibly be an exchange account.  There
	// could be an exchange *address* getting ready to become associated with an exchange account,
	// and in that case, it could get locked by this transaction.  However, it will then not be
	// allowed to be established later, by either SetValidation or CreateChildAccount.  Such accounts
	// will never become "exchange accounts" and exchanges will have to make a new address and
	// not transfer to and lock it.
	if exists {
		return errors.New("invalid TransferAndLock: cannot be sent to an existing account")
	}

	return nil
}

// Apply satisfies metatx.Transactable
func (tx *TransferAndLock) Apply(appInt interface{}) error {
	app := appInt.(*App)
	err := app.applyTxDetails(tx)
	if err != nil {
		return err
	}

	lockedBonusRateTable := eai.RateTable{}
	err = app.System(sv.LockedRateTableName, &lockedBonusRateTable)
	if err != nil {
		return err
	}

	source, _ := app.getAccount(tx.Source)
	// we know dest is a new account so WAA, WAA update and EAI update times are set properly
	dest, _ := app.getAccount(tx.Destination)
	dest.Balance = tx.Qty
	if source.RecourseSettings.Period != 0 {
		x := app.BlockTime().Add(source.RecourseSettings.Period)
		dest.Holds = append(dest.Holds, backing.Hold{
			Qty:    tx.Qty,
			Expiry: &x,
			Txhash: metatx.Hash(tx),
		})
	}
	dest.Lock = backing.NewLock(tx.Period, lockedBonusRateTable)

	dest.UpdateCurrencySeat(app.BlockTime())

	return app.UpdateState(func(stateI metast.State) (metast.State, error) {
		state := stateI.(*backing.State)

		state.Accounts[tx.Destination.String()] = dest
		state.Accounts[tx.Source.String()] = source

		return state, nil
	})
}

// GetSource implements Sourcer
func (tx *TransferAndLock) GetSource(*App) (address.Address, error) {
	return tx.Source, nil
}

// Withdrawal implements Withdrawer
func (tx *TransferAndLock) Withdrawal() math.Ndau {
	return tx.Qty
}

// GetSequence implements Sequencer
func (tx *TransferAndLock) GetSequence() uint64 {
	return tx.Sequence
}

// GetSignatures implements Signeder
func (tx *TransferAndLock) GetSignatures() []signature.Signature {
	return tx.Signatures
}

// ExtendSignatures implements Signable
func (tx *TransferAndLock) ExtendSignatures(sa []signature.Signature) {
	tx.Signatures = append(tx.Signatures, sa...)
}
