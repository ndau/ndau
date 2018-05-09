package ndau

import (
	nt "github.com/attic-labs/noms/go/types"
	util "github.com/oneiro-ndev/noms-util"
	"github.com/pkg/errors"
	"github.com/tendermint/abci/types"
)

var validatorKey = nt.String("validators")

func (app *App) validators() nt.Map {
	vals, hasVals := app.state.MaybeGet(validatorKey)
	if !hasVals {
		vals = nt.NewMap(app.db)
	}
	return vals.(nt.Map)
}

// Update the app's internal state with the given validator
func (app *App) updateValidator(v types.Validator) {
	logger := app.logger.With("method", "updateValidator")
	logger.Info("entered method", "Power", v.GetPower(), "PubKey", v.GetPubKey())
	validators := app.validators()
	pkBlob := util.Blob(app.db, v.GetPubKey())
	if v.Power == 0 {
		logger.Info("attempting to remove validator")
		validators = validators.Edit().Remove(pkBlob).Map()
	} else {
		logger.Info("attempting to update validator")
		powerBlob := util.Int(v.Power).ToBlob(app.db)
		validators = validators.Edit().Set(pkBlob, powerBlob).Map()
	}
	app.state = app.state.Edit().Set(validatorKey, validators).Map()

	// we only update the changes array after updating the tree
	app.ValUpdates = append(app.ValUpdates, v)
	logger.Info("exiting OK", "app.ValUpdates", app.ValUpdates)
}

// GetValidators returns a list of validators this app knows of
func (app *App) GetValidators() (validators []types.Validator, err error) {
	app.validators().IterAll(func(key, value nt.Value) {
		// this iterator interface doesn't allow for early failures,
		// so we need to just skip work if an error occurs
		if err != nil {
			return
		}
		pubKey, err := util.Unblob(key.(nt.Blob))
		err = errors.Wrap(err, "GetValidators found non-`nt.Blob` public key")
		if err != nil {
			return
		}
		power, err := util.IntFromBlob(value.(nt.Blob))
		err = errors.Wrap(err, "GetValidators found non-`nt.Blob` power")
		if err != nil {
			return
		}
		validators = append(validators, types.Validator{
			PubKey: pubKey,
			Power:  int64(power),
		})
	})
	return
}
