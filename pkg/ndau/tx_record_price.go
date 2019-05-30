package ndau

import (
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

// return a function intended to be run within app.UpdateState
//
// special case: if the input is negative, just use the existing value
func (app *App) updatePricesAndSIB(marketPrice pricecurve.Nanocent) func(stateI metast.State) (metast.State, error) {
	if marketPrice < 0 {
		marketPrice = app.GetState().(*backing.State).MarketPrice
	}
	return func(stateI metast.State) (metast.State, error) {
		sib, target, err := app.calculateCurrentSIB(marketPrice)
		if err != nil {
			return stateI, err
		}
		state := stateI.(*backing.State)
		state.SIB = sib
		state.MarketPrice = marketPrice
		state.TargetPrice = target

		return state, err
	}
}

// calculates the SIB implied by the market price given the current app state.
//
// It also returns the calculated target price.
func (app *App) calculateCurrentSIB(marketPrice pricecurve.Nanocent) (sib eai.Rate, targetPrice pricecurve.Nanocent, err error) {
	// compute the current target price
	state := app.GetState().(*backing.State)
	targetPrice, err = pricecurve.PriceAtUnit(state.TotalIssue)
	if err != nil {
		err = errors.Wrap(err, "computing target price")
		return
	}

	// get the script used to perform the calculation
	var sibScript wkt.Bytes
	err = app.System(sv.SIBScriptName, &sibScript)
	if err != nil {
		err = errors.Wrap(err, "fetching "+sv.SIBScriptName)
		return
	}
	if !IsChaincode(sibScript) {
		err = errors.New("sibScript appears not to be chaincode")
		return
	}

	// compute SIB
	vm, err := BuildVMForSIB(sibScript, uint64(targetPrice), uint64(marketPrice), app.BlockTime())
	if err != nil {
		err = errors.Wrap(err, "building vm for SIB calculation")
		return
	}

	err = vm.Run(nil)
	if err != nil {
		err = errors.Wrap(err, "computing SIB")
		return
	}

	top, err := vm.Stack().PopAsInt64()
	if err != nil {
		err = errors.Wrap(err, "retrieving SIB from VM")
		return
	}

	sib = eai.Rate(top)
	return
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

	return app.UpdateState(app.updatePricesAndSIB(tx.MarketPrice))
}

// GetSource implements Sourcer
func (tx *RecordPrice) GetSource(app *App) (addr address.Address, err error) {
	err = app.System(sv.RecordPriceAddressName, &addr)
	if err != nil {
		return
	}
	if addr.Revalidate() != nil {
		err = fmt.Errorf(
			"%s sysvar not set; RecordPrice therefore disallowed",
			sv.RecordPriceAddressName,
		)
		return
	}
	return
}

// GetSequence implements Sequencer
func (tx *RecordPrice) GetSequence() uint64 {
	return tx.Sequence
}

// GetSignatures implements Signeder
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
