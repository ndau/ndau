package ndau

import (
	metatx "github.com/oneiro-ndev/metanode/pkg/meta/transaction"
	"github.com/oneiro-ndev/msgp-well-known-types/wkt"
	sv "github.com/oneiro-ndev/ndau/pkg/ndau/system_vars"
	math "github.com/oneiro-ndev/ndaumath/pkg/types"
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

	err = vm.Run(false)
	if err != nil {
		return 0, errors.Wrap(err, "tx fee script")
	}

	vmReturn, err := vm.Stack().PopAsInt64()
	if err != nil {
		return 0, errors.Wrap(err, "tx fee script exited without numeric top value")
	}
	return math.Ndau(vmReturn), nil
}
