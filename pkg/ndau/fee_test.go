package ndau

import (
	"testing"

	"github.com/oneiro-ndev/metanode/pkg/meta/app/code"
	"github.com/oneiro-ndev/ndau/pkg/ndau/backing"
	"github.com/oneiro-ndev/ndaumath/pkg/constants"
	"github.com/stretchr/testify/require"
)

func TestTxFeesAreAddedToPendingNodeRewards(t *testing.T) {
	app, private := initAppTx(t)

	pnr := app.GetState().(*backing.State).PendingNodeReward

	tx := NewTransfer(sourceAddress, destAddress, 1*constants.NapuPerNdau, 1, private)
	resp := deliverTxWithTxFee(t, app, tx)
	require.Equal(t, code.OK, code.ReturnCode(resp.Code))

	newPNR := app.GetState().(*backing.State).PendingNodeReward
	require.Equal(t, pnr+1, newPNR, "testing with tx fee of 1 should add 1 to PNR after single tx")
}
