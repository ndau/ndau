package ndau

import (
	metast "github.com/oneiro-ndev/metanode/pkg/meta/state"
	"github.com/oneiro-ndev/ndau/pkg/ndau/backing"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	"github.com/oneiro-ndev/ndaumath/pkg/eai"
	"github.com/oneiro-ndev/ndaumath/pkg/signature"
	sv "github.com/oneiro-ndev/system_vars/pkg/system_vars"
	"github.com/pkg/errors"
)

// GetAccountAddresses returns the account addresses associated with this transaction type.
func (tx *Lock) GetAccountAddresses() []string {
	return []string{tx.Target.String()}
}

// Validate implements metatx.Transactable
func (tx *Lock) Validate(appI interface{}) error {
	app := appI.(*App)

	accountData, hasAccount, _, err := app.getTxAccount(tx)
	if err != nil {
		return err
	}

	if !hasAccount {
		return errors.New("No such account")
	}

	if accountData.Lock != nil {
		if accountData.Lock.UnlocksOn == nil {
			// if not notified, lock is valid if its period >= the current period
			if tx.Period < accountData.Lock.NoticePeriod {
				return errors.New("Locked, non-notified accounts may be relocked only with periods >= their current")
			}
		} else {
			// if notified, lock is valid if it expires after the current unlock time
			lockExpiry := app.blockTime.Add(tx.Period)
			uo := *accountData.Lock.UnlocksOn
			if lockExpiry.Compare(uo) < 0 {
				return errors.New("Locked, notified accounts may be relocked only when new lock min expiry >= current unlock time")
			}
		}
	}

	// Ensure that this is not an exchange account, as they are not allowed to be locked.
	accountAttributes := sv.AccountAttributes{}
	err = app.System(sv.AccountAttributesName, &accountAttributes)
	if err != nil {
		return err
	}
	var progenitor *address.Address
	target, _ := app.getAccount(tx.Target)
	if target.Progenitor == nil {
		progenitor = &tx.Target
	} else {
		progenitor = target.Progenitor
	}
	if attributes, ok := accountAttributes[progenitor.String()]; ok {
		if _, ok := attributes[sv.AccountAttributeExchange]; ok {
			return errors.New("Cannot lock exchange accounts")
		}
	}

	return nil
}

// Apply implements metatx.Transactable
func (tx *Lock) Apply(appI interface{}) error {
	app := appI.(*App)
	err := app.applyTxDetails(tx)
	if err != nil {
		return err
	}

	lockedBonusRateTable := eai.RateTable{}
	err = app.System(sv.LockedRateTableName, &lockedBonusRateTable)
	if err != nil {
		return err
	}

	return app.UpdateState(func(stateI metast.State) (metast.State, error) {
		state := stateI.(*backing.State)
		accountData, _ := app.getAccount(tx.Target)

		accountData.Lock = backing.NewLock(tx.Period, lockedBonusRateTable)

		state.Accounts[tx.Target.String()] = accountData
		return state, nil
	})
}

// GetSource implements sourcer
func (tx *Lock) GetSource(*App) (address.Address, error) {
	return tx.Target, nil
}

// GetSequence implements sequencer
func (tx *Lock) GetSequence() uint64 {
	return tx.Sequence
}

// GetSignatures implements signeder
func (tx *Lock) GetSignatures() []signature.Signature {
	return tx.Signatures
}

// ExtendSignatures implements Signable
func (tx *Lock) ExtendSignatures(sa []signature.Signature) {
	tx.Signatures = append(tx.Signatures, sa...)
}
