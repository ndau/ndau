package ndau

import (
	"testing"
	"time"

	"github.com/oneiro-ndev/metanode/pkg/meta.app/code"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	math "github.com/oneiro-ndev/ndaumath/pkg/types"
	"github.com/oneiro-ndev/ndaunode/pkg/ndau/backing"
	"github.com/stretchr/testify/require"
)

func TestCEPStoresPendingEscrowChange(t *testing.T) {
	app, private := initAppTx(t)
	acct := app.GetState().(*backing.State).Accounts[source]
	require.Equal(t, math.Duration(0), acct.EscrowSettings.Duration)

	const newDuration = math.Duration(1 * math.Day)

	addr, err := address.Validate(source)
	require.NoError(t, err)

	cep, err := NewChangeEscrowPeriod(addr, newDuration, acct.Sequence+1, private)
	require.NoError(t, err)

	ts := time.Now()
	mts, err := math.TimestampFrom(ts)
	require.NoError(t, err)

	resp := deliverTrAt(t, app, &cep, ts.Unix())
	require.Equal(t, code.OK, code.ReturnCode(resp.Code))

	// the state of UpdateBalance is formally undefined at this point:
	// it's legit for implementations to notice that the current UpdateDuration
	// is 0, and to update the escrow period immediately. It's also legit
	// for them to wait for an UpdateEscrow call.

	// update the acct struct
	acct = app.GetState().(*backing.State).Accounts[source]
	acct.UpdateEscrow(mts)

	require.Equal(t, newDuration, acct.EscrowSettings.Duration)
	require.Nil(t, acct.EscrowSettings.Next)
	require.Nil(t, acct.EscrowSettings.ChangesAt)
}
