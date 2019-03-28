package ndau

import (
	metatx "github.com/oneiro-ndev/metanode/pkg/meta/transaction"
	"github.com/oneiro-ndev/msgp-well-known-types/wkt"
	"github.com/oneiro-ndev/ndau/pkg/ndau/backing"
	"github.com/oneiro-ndev/ndaumath/pkg/constants"
	"github.com/oneiro-ndev/ndaumath/pkg/signed"
	math "github.com/oneiro-ndev/ndaumath/pkg/types"
	sv "github.com/oneiro-ndev/system_vars/pkg/system_vars"
	"github.com/pkg/errors"
)

func (app *App) calculateTxFee(tx metatx.Transactable) (math.Ndau, error) {
	var script wkt.Bytes
	err := app.System(sv.TxFeeScriptName, &script)
	if err != nil {
		return 0, errors.Wrap(err, "fetching TxFeeScript system variable")
	}

	vm, err := BuildVMForTxFees(script, tx, app.blockTime)
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

func (app *App) calculateSIB(tx NTransactable) (math.Ndau, error) {
	if w, ok := tx.(withdrawer); ok {
		sibRate := app.GetState().(*backing.State).SIB
		if sibRate > 0 {
			source, err := tx.GetSource(app)
			if err != nil {
				return 0, errors.Wrap(err, "getting tx source")
			}
			isExchangeAccount, err := app.GetState().(*backing.State).AccountHasAttribute(source, sv.AccountAttributeExchange)
			if err != nil {
				return 0, errors.Wrap(err, "determing whether tx source is exchange account")
			}
			if !isExchangeAccount {
				sib, err := signed.MulDiv(
					int64(w.Withdrawal()),
					int64(sibRate),
					constants.RateDenominator,
				)
				return math.Ndau(sib), errors.Wrap(err, "calculating SIB")
			}
		}
	}
	return 0, nil
}
