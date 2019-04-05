package ndau

import (
	"testing"
	"time"

	"github.com/oneiro-ndev/metanode/pkg/meta/app/code"
	"github.com/oneiro-ndev/ndau/pkg/ndau/backing"
	math "github.com/oneiro-ndev/ndaumath/pkg/types"
	"github.com/stretchr/testify/require"
)

func TestCSPStoresPendingEscrowChange(t *testing.T) {
	app, private := initAppTx(t)
	acct := app.GetState().(*backing.State).Accounts[source]
	require.Equal(t, math.Duration(0), acct.SettlementSettings.Period)

	const newDuration = math.Duration(1 * math.Day)

	addr := sourceAddress

	cep := NewChangeRecoursePeriod(addr, newDuration, acct.Sequence+1, private)

	ts, err := math.TimestampFrom(time.Now())
	require.NoError(t, err)

	resp := deliverTxAt(t, app, cep, ts)
	require.Equal(t, code.OK, code.ReturnCode(resp.Code))

	// the state of UpdateBalance is formally undefined at this point:
	// it's legit for implementations to notice that the current UpdateDuration
	// is 0, and to update the escrow period immediately. It's also legit
	// for them to wait for an UpdateEscrow call.

	// update the acct struct
	acct = app.GetState().(*backing.State).Accounts[source]
	acct.UpdateSettlements(ts)

	require.Equal(t, newDuration, acct.SettlementSettings.Period)
	require.Nil(t, acct.SettlementSettings.Next)
	require.Nil(t, acct.SettlementSettings.ChangesAt)
}

func TestChangeSettlementPeriodDeductsTxFee(t *testing.T) {
	app, private := initAppTx(t)
	modify(t, source, app, func(ad *backing.AccountData) {
		ad.Balance = 1
	})

	addr := sourceAddress

	const newDuration = math.Duration(1 * math.Day)

	for i := 0; i < 2; i++ {
		tx := NewChangeRecoursePeriod(
			addr,
			newDuration,
			uint64(i)+1,
			private,
		)

		resp := deliverTxWithTxFee(t, app, tx)

		var expect code.ReturnCode
		if i == 0 {
			expect = code.OK
		} else {
			expect = code.InvalidTransaction
		}
		require.Equal(t, expect, code.ReturnCode(resp.Code))
	}
}
