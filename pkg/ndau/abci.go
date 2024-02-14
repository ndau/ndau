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
	metast "github.com/ndau/metanode/pkg/meta/state"
	"github.com/ndau/msgp-well-known-types/wkt"
	"github.com/ndau/ndau/pkg/ndau/backing"
	sv "github.com/ndau/system_vars/pkg/system_vars"
	log "github.com/sirupsen/logrus"
	abci "github.com/tendermint/tendermint/abci/types"
)

// BeginBlock overrides the metanode BeginBlock ABCI message handler.
//
// If a quit is pending, the application (and the ndaunode executable) exits.
// Otherwise, just uses the default handler.
func (app *App) BeginBlock(req abci.RequestBeginBlock) abci.ResponseBeginBlock {
	if app.quitPending {
		quit()
	}
	return app.App.BeginBlock(req)
}

// EndBlock updates the validator set, compositing its behavior with metanode's
func (app *App) EndBlock(req abci.RequestEndBlock) abci.ResponseEndBlock {
	reb := app.App.EndBlock(req)

	logger := app.DecoratedLogger().WithFields(log.Fields{
		"method": "ndau.App.EndBlock",
	})

	// if sv.NodeMaxValidators is set, then the top n nodes by goodness
	// must be assigned voting power proportional to their goodness.
	// All other nodes must be assigned 0 voting power.
	var maxValidators wkt.Uint64
	err := app.System(sv.NodeMaxValidators, &maxValidators)
	if err == nil && app.IsFeatureActive("MaxValidatorsOn") {
		logger = logger.WithField("endblock.max_validators", maxValidators)
		// get goodnesses
		gs, _ := nodeGoodnesses(app)
		// filter down the top n
		gs = topNGoodnesses(gs, int(maxValidators))

		logger = logger.WithField("endblock.len_gs", len(gs))

		// for each remaining goodness, create a corresponding validator update
		state := app.GetState().(*backing.State)
		for _, g := range gs {
			vu, err := validatorUpdateFor(state, g.addr)
			if err != nil {
				logger.WithError(err).WithFields(log.Fields{
					"node":     g.addr,
					"goodness": g.goodness,
				}).Error("creating validator update")
				continue
			}
			vu.Power = int64(g.goodness)
			reb.ValidatorUpdates = append(reb.ValidatorUpdates, *vu)
			// If we've told tendermint to remove the validator by setting its
			// voting power to 0, then we need to remove the validator from our
			// own list of nodes. This is exactly what the UnregisterNode
			// transaction does. It's not worth trying to batch the updates
			// since most of the time only one validator will be deleted in a block.
			if vu.Power == 0 {
				app.UpdateState(func(stateI metast.State) (metast.State, error) {
					state := stateI.(*backing.State)
					delete(state.Nodes, g.addr)
					return state, nil
				})
			}
		}
		logger.WithField("endblock.len_vus", len(reb.ValidatorUpdates))
		logger.Info("updated node validation power")
	} else {
		logger.WithError(err).Info("could not get max validators sysvar; skipping updates")
	}

	return reb
}
