package ndau

import (
	"encoding/base64"

	"github.com/oneiro-ndev/msgp-well-known-types/wkt"
	"github.com/oneiro-ndev/ndau/pkg/ndau/backing"
	"github.com/oneiro-ndev/ndaumath/pkg/eai"
	sv "github.com/oneiro-ndev/system_vars/pkg/system_vars"
	"github.com/pkg/errors"
)

func (app *App) calculateExchangeEAIRate(exchangeAccount backing.AccountData) (eai.Rate, error) {
	var script wkt.Bytes
	exists, err := app.SystemOptional(sv.ExchangeEAIScriptName, &script)
	if err != nil {
		if exists {
			// Some critical error occurred fetching the system variable.
			return 0, errors.Wrap(err, "Could not fetch ExchangeEAIScript system variable")
		}
		// The system variable doesn't exist, use the default.
		script, err = base64.StdEncoding.DecodeString(sv.ExchangeEAIScriptDefault)
		if err != nil {
			return 0, errors.Wrap(err, "Could not decode ExchangeEAIScriptDefault")
		}
	}

	vm, err := BuildVMForExchangeEAI(script, exchangeAccount, app.GetState().(*backing.State).SIB)
	if err != nil {
		return 0, errors.Wrap(err, "Could not build vm for exchange EAI script")
	}

	err = vm.Run(nil)
	if err != nil {
		return 0, errors.Wrap(err, "Could not run exchange EAI script")
	}

	vmReturn, err := vm.Stack().PopAsInt64()
	if err != nil {
		return 0, errors.Wrap(err, "Exchange EAI script exited without numeric top value")
	}

	return eai.Rate(vmReturn), nil
}
