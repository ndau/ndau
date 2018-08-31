package ndau

import (
	"testing"
	"time"

	"github.com/oneiro-ndev/metanode/pkg/meta/app/code"
	"github.com/oneiro-ndev/ndau/pkg/ndau/backing"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	math "github.com/oneiro-ndev/ndaumath/pkg/types"
	"github.com/oneiro-ndev/signature/pkg/signature"
	"github.com/stretchr/testify/require"
)

func TestCSPStoresPendingEscrowChange(t *testing.T) {
	app, private := initAppTx(t)
	acct := app.GetState().(*backing.State).Accounts[source]
	require.Equal(t, math.Duration(0), acct.SettlementSettings.Period)

	const newDuration = math.Duration(1 * math.Day)

	addr, err := address.Validate(source)
	require.NoError(t, err)

	cep, err := NewChangeSettlementPeriod(addr, newDuration, acct.Sequence+1, []signature.PrivateKey{private})
	require.NoError(t, err)

	ts, err := math.TimestampFrom(time.Now())
	require.NoError(t, err)

	resp := deliverTrAt(t, app, &cep, ts)
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

func TestChangeSettlemntPeriodDeductsTxFee(t *testing.T) {
	app, private := initAppTx(t)
	modify(t, source, app, func(ad *backing.AccountData) {
		ad.Balance = 1
	})

	addr, err := address.Validate(source)
	require.NoError(t, err)

	const newDuration = math.Duration(1 * math.Day)

	for i := 0; i < 2; i++ {
		tx, err := NewChangeSettlementPeriod(
			addr,
			newDuration,
			uint64(i)+1,
			[]signature.PrivateKey{private},
		)
		require.NoError(t, err)

		resp := deliverTrWithTxFee(t, app, &tx)

		var expect code.ReturnCode
		if i == 0 {
			expect = code.OK
		} else {
			expect = code.InvalidTransaction
		}
		require.Equal(t, expect, code.ReturnCode(resp.Code))
	}
}
