package ndau

import (
	"fmt"
	"os"

	"github.com/oneiro-ndev/ndaumath/pkg/address"
	"github.com/oneiro-ndev/ndaumath/pkg/signature"
	sv "github.com/oneiro-ndev/system_vars/pkg/system_vars"
)

// ChangeSchemaExitCode is returned when the ndaunode exits due to ChangeSchema
const ChangeSchemaExitCode = 0xdead

// Validate implements metatx.Transactable
func (tx *ChangeSchema) Validate(appI interface{}) error {
	app := appI.(*App)

	_, _, _, err := app.getTxAccount(tx)

	return err
}

// Apply implements metatx.Transactable
func (tx *ChangeSchema) Apply(appI interface{}) error {
	app := appI.(*App)
	err := app.applyTxDetails(tx)
	if err != nil {
		return err
	}

	app.DecoratedTxLogger(tx).Warn("System going down due to ChangeSchema tx")
	os.Exit(ChangeSchemaExitCode)
	return nil
}

// GetSource implements sourcer
func (tx *ChangeSchema) GetSource(app *App) (addr address.Address, err error) {
	err = app.System(sv.ChangeSchemaAddressName, &addr)
	if err != nil {
		return
	}
	if addr.Revalidate() != nil {
		err = fmt.Errorf(
			"%s sysvar not set; ChangeSchema therefore disallowed",
			sv.RecordPriceAddressName,
		)
		return
	}
	return
}

// GetSequence implements sequencer
func (tx *ChangeSchema) GetSequence() uint64 {
	return tx.Sequence
}

// GetSignatures implements signeder
func (tx *ChangeSchema) GetSignatures() []signature.Signature {
	return tx.Signatures
}

// ExtendSignatures implements Signable
func (tx *ChangeSchema) ExtendSignatures(sa []signature.Signature) {
	tx.Signatures = append(tx.Signatures, sa...)
}

// GetAccountAddresses returns the account addresses associated with this transaction type.
func (tx *ChangeSchema) GetAccountAddresses() []string {
	return []string{}
}
