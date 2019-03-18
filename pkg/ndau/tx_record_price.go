package ndau

import (
	"encoding/base64"
	"fmt"

	metast "github.com/oneiro-ndev/metanode/pkg/meta/state"
	"github.com/oneiro-ndev/msgp-well-known-types/wkt"
	"github.com/oneiro-ndev/ndau/pkg/ndau/backing"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	"github.com/oneiro-ndev/ndaumath/pkg/eai"
	"github.com/oneiro-ndev/ndaumath/pkg/pricecurve"
	"github.com/oneiro-ndev/ndaumath/pkg/signature"
	sv "github.com/oneiro-ndev/system_vars/pkg/system_vars"
	"github.com/pkg/errors"
)

// CalculateSIB calculates the SIB implied by the market price given the current app state.
func (tx *RecordPrice) CalculateSIB(app *App) (eai.Rate, error) {
	// compute the current target price
	state := app.GetState().(*backing.State)
	targetPrice, err := pricecurve.PriceAtUnit(state.TotalIssue)
	if err != nil {
		return 0, errors.Wrap(err, "computing target price")
	}

	// get the script used to perform the calculation
	var sibScript wkt.Bytes
	exists, err := app.SystemOptional(sv.SIBScriptName, &sibScript)
	if err != nil {
		return 0, errors.Wrap(err, "fetching "+sv.SIBScriptName)
	}
	if !exists {
		sibScript, err = base64.StdEncoding.DecodeString(sv.SIBScriptDefault)
		if err != nil {
			return 0, errors.Wrap(err, "decoding sv.SIBScriptDefault")
		}
	}
	if !IsChaincode(sibScript) {
		return 0, errors.New("sibScript appears not to be chaincode")
	}

	// compute SIB
	vm, err := BuildVMForSIB(sibScript, uint64(targetPrice), uint64(tx.MarketPrice), app.blockTime)
	if err != nil {
		return 0, errors.Wrap(err, "building vm for SIB calculation")
	}

	err = vm.Run(nil)
	if err != nil {
		return 0, errors.Wrap(err, "computing SIB")
	}

	top, err := vm.Stack().PopAsInt64()
	if err != nil {
		return 0, errors.Wrap(err, "retrieving SIB from VM")
	}

	return eai.Rate(top), nil
}

// Validate implements metatx.Transactable
func (tx *RecordPrice) Validate(appI interface{}) error {
	app := appI.(*App)

	if tx.MarketPrice <= 0 {
		return errors.New("RecordPrice market price may not be <= 0")
	}

	_, _, _, err := app.getTxAccount(tx)

	return err
}

// Apply implements metatx.Transactable
func (tx *RecordPrice) Apply(appI interface{}) error {
	app := appI.(*App)
	err := app.applyTxDetails(tx)
	if err != nil {
		return err
	}

	return app.UpdateState(func(stateI metast.State) (metast.State, error) {
		sib, err := tx.CalculateSIB(app)
		if err != nil {
			return stateI, err
		}
		state := stateI.(*backing.State)
		state.SIB = sib

		return state, err
	})
}

// GetSource implements sourcer
func (tx *RecordPrice) GetSource(app *App) (addr address.Address, err error) {
	err = app.System(sv.RecordPriceAddressName, &addr)
	if err != nil {
		return
	}
	if addr.String() == "" {
		err = fmt.Errorf(
			"%s sysvar not set; RecordPrice therefore disallowed",
			sv.RecordPriceAddressName,
		)
		return
	}
	return
}

// GetSequence implements sequencer
func (tx *RecordPrice) GetSequence() uint64 {
	return tx.Sequence
}

// GetSignatures implements signeder
func (tx *RecordPrice) GetSignatures() []signature.Signature {
	return tx.Signatures
}

// ExtendSignatures implements Signable
func (tx *RecordPrice) ExtendSignatures(sa []signature.Signature) {
	tx.Signatures = append(tx.Signatures, sa...)
}

// GetAccountAddresses returns the account addresses associated with this transaction type.
func (tx *RecordPrice) GetAccountAddresses() []string {
	return []string{}
}
