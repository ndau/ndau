package ndau

import (
	"errors"

	metast "github.com/oneiro-ndev/metanode/pkg/meta/state"
	"github.com/oneiro-ndev/ndau/pkg/ndau/backing"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	"github.com/oneiro-ndev/ndaumath/pkg/eai"
	"github.com/oneiro-ndev/ndaumath/pkg/signature"
	math "github.com/oneiro-ndev/ndaumath/pkg/types"
	sv "github.com/oneiro-ndev/system_vars/pkg/system_vars"
)

// Validate implements metatx.Transactable
func (tx *Burn) Validate(appI interface{}) error {
	app := appI.(*App)

	if tx.Qty <= 0 {
		return errors.New("burn qty must be positive")
	}

	acctData, _, _, err := app.getTxAccount(tx)
	if err != nil {
		return err
	}

	if acctData.IsLocked(app.BlockTime()) {
		return errors.New("burn from locked accounts prohibited")
	}

	return nil
}

// Apply implements metatx.Transactable
func (tx *Burn) Apply(appI interface{}) error {
	app := appI.(*App)

	lockedBonusRateTable := eai.RateTable{}
	err := app.System(sv.LockedRateTableName, &lockedBonusRateTable)
	if err != nil {
		return err
	}

	return app.UpdateState(app.applyTxDetails(tx), func(stI metast.State) (metast.State, error) {
		st := stI.(*backing.State)
		st.TotalBurned += tx.Qty
		return st, nil
	})
}

// GetSource implements Sourcer
func (tx *Burn) GetSource(*App) (address.Address, error) {
	return tx.Target, nil
}

// Withdrawal implements Withdrawer
func (tx *Burn) Withdrawal() math.Ndau {
	return tx.Qty
}

// GetSequence implements Sequencer
func (tx *Burn) GetSequence() uint64 {
	return tx.Sequence
}

// GetSignatures implements Signeder
func (tx *Burn) GetSignatures() []signature.Signature {
	return tx.Signatures
}

// ExtendSignatures implements Signable
func (tx *Burn) ExtendSignatures(sa []signature.Signature) {
	tx.Signatures = append(tx.Signatures, sa...)
}
