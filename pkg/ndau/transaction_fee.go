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
	metatx "github.com/ndau/metanode/pkg/meta/transaction"
	"github.com/ndau/msgp-well-known-types/wkt"
	"github.com/ndau/ndau/pkg/ndau/backing"
	"github.com/ndau/ndaumath/pkg/constants"
	"github.com/ndau/ndaumath/pkg/signed"
	math "github.com/ndau/ndaumath/pkg/types"
	sv "github.com/ndau/system_vars/pkg/system_vars"
	"github.com/pkg/errors"
)

func (app *App) calculateTxFee(tx metatx.Transactable) (math.Ndau, error) {
	var script wkt.Bytes
	err := app.System(sv.TxFeeScriptName, &script)
	if err != nil {
		return 0, errors.Wrap(err, "fetching TxFeeScript system variable")
	}

	vm, err := BuildVMForTxFees(script, tx, app.BlockTime())
	if err != nil {
		return 0, errors.Wrap(err, "couldn't build vm for tx fee script")
	}

	err = vm.Run(nil)
	if err != nil {
		return 0, errors.Wrap(err, "tx fee script")
	}

	vmReturn, err := vm.Stack().PopAsInt64()
	if err != nil {
		return 0, errors.Wrap(err, "tx fee script exited without numeric top value")
	}
	return math.Ndau(vmReturn), nil
}

// Change in SIB application rules. Previously SIB was imposed except if the source was an authorized
// exchange account. Now SIB will be imposed only if the source is not an authorized exchange account
// and the destination is an authorized exchange account.

// Addresses created with an "ndx" prefix are now always treated as exchange addresses for the purpose
// of calculating SIB, regardless of whether they're actually marked as authorized exchange addresses.

func (app *App) calculateSIB(tx NTransactable) (math.Ndau, error) {
	if w, ok := tx.(Withdrawer); ok {
		sibRate := app.GetState().(*backing.State).SIB
		if sibRate > 0 {
			source, err := tx.GetSource(app)
			if err != nil {
				return 0, errors.Wrap(err, "getting tx source")
			}
			isSourceExchangeAccount, err := app.GetState().(*backing.State).AccountHasAttribute(source, sv.AccountAttributeExchange)
			if err != nil {
				return 0, errors.Wrap(err, "determing whether tx source is exchange account")
			}
			// Check for exchange-format addresses that haven't been marked as children of authorized exchanges
			if source.String()[:3] == "ndx" {
				isSourceExchangeAccount = true
			}
			if !isSourceExchangeAccount {
				// JSG only apply SIB if src is not exchange and dest is exchange
				if app.IsFeatureActive("NewSIBRules") {
					if d, ok := tx.(HasDestination); ok {
						dest, err := d.GetDestination(app)
						if err != nil {
							return 0, errors.Wrap(err, "getting tx destination")
						}
						isDestinationExchangeAccount, err := app.GetState().(*backing.State).AccountHasAttribute(dest, sv.AccountAttributeExchange)
						if err != nil {
							return 0, errors.Wrap(err, "determining whether tx destination is an exchange account")
						}

						// Check for exchange-format destinations that haven't been set up with CreateChildAccount yet,
						// to prevent such addresses from avoiding SIB on incoming transfers.

						// Fix for v1.5.2 - this feature did not have a height gate specified. There were several such
						// transfers made (nda->ndx with no SIB) before this fix, so the current image is no longer able
						// to catch up from genesis.

						if app.IsFeatureActive("ndxAreAllExchange") {
							if dest.String()[:3] == "ndx" {
								isDestinationExchangeAccount = true
							}
						}
						if isDestinationExchangeAccount {
							sib, err := signed.MulDiv(
								int64(w.Withdrawal()),
								int64(sibRate),
								constants.RateDenominator,
							)
							return math.Ndau(sib), errors.Wrap(err, "calculating SIB")
						}
					}
				} else {
					sib, err := signed.MulDiv(
						int64(w.Withdrawal()),
						int64(sibRate),
						constants.RateDenominator,
					)
					return math.Ndau(sib), errors.Wrap(err, "calculating SIB")
				}
			}
		}
	}
	return 0, nil
}
