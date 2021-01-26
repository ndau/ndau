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
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	metast "github.com/ndau/metanode/pkg/meta/state"
	"github.com/ndau/ndau/pkg/ndau/backing"
	"github.com/ndau/ndaumath/pkg/address"
	"github.com/ndau/ndaumath/pkg/signature"
	math "github.com/ndau/ndaumath/pkg/types"
	sv "github.com/ndau/system_vars/pkg/system_vars"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

// Validate implements metatx.Transactable
func (tx *NominateNodeReward) Validate(appI interface{}) error {
	app := appI.(*App)

	_, _, _, err := app.getTxAccount(tx)
	if err != nil {
		return err
	}

	state := app.GetState().(*backing.State)

	// enough time must have elapsed
	minDuration := math.Duration(0)
	err = app.System(sv.MinDurationBetweenNodeRewardNominationsName, &minDuration)
	if err != nil {
		return errors.Wrap(err, "getting min duration system variable")
	}
	if app.BlockTime().Since(state.LastNodeRewardNomination) < minDuration {
		return fmt.Errorf(
			"not enough time since last NNR. need %s, have %s",
			minDuration,
			app.BlockTime().Since(state.LastNodeRewardNomination),
		)
	}

	return err
}

// Apply implements metatx.Transactable
func (tx *NominateNodeReward) Apply(appI interface{}) error {
	app := appI.(*App)

	return app.UpdateState(app.applyTxDetails(tx), func(stateI metast.State) (metast.State, error) {
		state := stateI.(*backing.State)

		state.LastNodeRewardNomination = app.BlockTime()
		state.UnclaimedNodeReward = state.PendingNodeReward
		state.PendingNodeReward = 0

		nrw, err := app.SelectByGoodness(uint64(tx.Random))
		if err != nil {
			return state, err
		}

		wh := ""
		if app.config.NodeRewardWebhook != nil {
			wh = *app.config.NodeRewardWebhook
		}
		logger := app.DecoratedTxLogger(tx).WithFields(log.Fields{
			"webhook": wh,
		})
		logger.Info("launching zzzzz callWinnerWebhook goroutine")

		go app.callWinnerWebhook(tx, nrw, logger)
		state.NodeRewardWinner = &nrw
		return state, err
	})
}

// this is a noop normally, but tests can use it to synchronize a waitgroup
// once the webhook completes
var whDone func()

func init() {
	whDone = func() {}
}

func (app *App) callWinnerWebhook(tx *NominateNodeReward, winner address.Address, logger *log.Entry) {
	defer whDone()

	if app.config.NodeRewardWebhook == nil {
		return
	}

	if app.config.NodeRewardWebhookDelay != nil {
		time.Sleep(time.Duration(*app.config.NodeRewardWebhookDelay*rand.Float64()) * time.Second)
	}

	logger = logger.WithField("method", "callWinnerWebhook")

	body := struct {
		Random int64  `json:"random"`
		Winner string `json:"winner"`
	}{
		Random: tx.Random,
		Winner: winner.String(),
	}

	buff := new(bytes.Buffer)
	err := json.NewEncoder(buff).Encode(body)
	if err != nil {
		logger.WithError(err).Error("failed to encode body as json")
		return
	}

	resp, err := (&http.Client{Timeout: 30 * time.Second}).Post(
		*app.config.NodeRewardWebhook,
		"application/json",
		buff,
	)
	if err != nil {
		logger.WithError(err).Error("failed to send webhook request")
		return
	}
	resp.Body.Close()

	logger.Info("successfully posted node reward winner to webhook")
}

// GetSource implements Sourcer
func (tx *NominateNodeReward) GetSource(app *App) (addr address.Address, err error) {
	err = app.System(sv.NominateNodeRewardAddressName, &addr)
	return
}

// GetSequence implements Sequencer
func (tx *NominateNodeReward) GetSequence() uint64 {
	return tx.Sequence
}

// GetSignatures implements Signeder
func (tx *NominateNodeReward) GetSignatures() []signature.Signature {
	return tx.Signatures
}

// ExtendSignatures implements Signable
func (tx *NominateNodeReward) ExtendSignatures(sa []signature.Signature) {
	tx.Signatures = append(tx.Signatures, sa...)
}
