package ndau

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	metast "github.com/oneiro-ndev/metanode/pkg/meta/state"
	"github.com/oneiro-ndev/ndau/pkg/ndau/backing"
	sv "github.com/oneiro-ndev/ndau/pkg/ndau/system_vars"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	math "github.com/oneiro-ndev/ndaumath/pkg/types"
	"github.com/oneiro-ndev/signature/pkg/signature"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

// SignableBytes implements Transactable
func (tx *NominateNodeReward) SignableBytes() []byte {
	bytes := make([]byte, 0, 8+8)
	bytes = appendUint64(bytes, tx.Random)
	bytes = appendUint64(bytes, tx.Sequence)
	return bytes
}

// NewNominateNodeReward constructs a NominateNodeReward transactable.
//
// The caller must ensure that `private` corresponds to a public key listed
// in the `NominateNodeRewardKeys` system variable.
func NewNominateNodeReward(
	random uint64,
	sequence uint64,
	keys []signature.PrivateKey,
) (tx NominateNodeReward) {
	tx = NominateNodeReward{
		Random:   random,
		Sequence: sequence,
	}
	for _, key := range keys {
		tx.Signatures = append(tx.Signatures, key.Sign(tx.SignableBytes()))
	}
	return tx
}

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
	if app.blockTime.Since(state.LastNodeRewardNomination) < minDuration {
		return fmt.Errorf(
			"not enough time since last NNR. need %s, have %s",
			minDuration,
			app.blockTime.Since(state.LastNodeRewardNomination),
		)
	}

	return err
}

// Apply implements metatx.Transactable
func (tx *NominateNodeReward) Apply(appI interface{}) error {
	app := appI.(*App)
	err := app.applyTxDetails(tx)
	if err != nil {
		return err
	}

	return app.UpdateState(func(stateI metast.State) (metast.State, error) {
		var err error
		state := stateI.(*backing.State)

		state.LastNodeRewardNomination = app.blockTime
		state.UnclaimedNodeReward = state.PendingNodeReward
		state.PendingNodeReward = 0

		winner, err := app.SelectByGoodness(tx.Random)
		if err != nil {
			return state, err
		}
		go app.callWinnerWebhook(tx, winner)
		return state, err
	})
}

func (app *App) callWinnerWebhook(tx *NominateNodeReward, winner address.Address) {
	if app.config.NodeRewardWebhook == nil {
		return
	}

	logger := app.DecoratedTxLogger(tx).WithFields(log.Fields{
		"method":  "callWinnerWebhook",
		"webhook": *app.config.NodeRewardWebhook,
	})

	body := struct {
		Random uint64 `json:"random"`
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

	resp, err := http.Post(
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

// GetSource implements sourcer
func (tx *NominateNodeReward) GetSource(app *App) (addr address.Address, err error) {
	err = app.System(sv.NominateNodeRewardAddressName, &addr)
	return
}

// GetSequence implements sequencer
func (tx *NominateNodeReward) GetSequence() uint64 {
	return tx.Sequence
}

// GetSignatures implements signeder
func (tx *NominateNodeReward) GetSignatures() []signature.Signature {
	return tx.Signatures
}
