package ndau

import (
	"bytes"
	"encoding/json"
	"fmt"

	abci "github.com/tendermint/tendermint/abci/types"

	meta "github.com/oneiro-ndev/metanode/pkg/meta/app"
	"github.com/oneiro-ndev/ndau/pkg/ndau/backing"
	"github.com/oneiro-ndev/ndau/pkg/query"
	"github.com/oneiro-ndev/ndau/pkg/version"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	"github.com/oneiro-ndev/ndaumath/pkg/types"
)

func init() {
	meta.RegisterQueryHandler(query.AccountEndpoint, accountQuery)
	meta.RegisterQueryHandler(query.SummaryEndpoint, summaryQuery)
	meta.RegisterQueryHandler(query.VersionEndpoint, versionQuery)
	meta.RegisterQueryHandler(query.SysvarsEndpoint, sysvarsQuery)
}

func accountQuery(appI interface{}, request abci.RequestQuery, response *abci.ResponseQuery) {
	app := appI.(*App)

	address, err := address.Validate(string(request.GetData()))
	if err != nil {
		app.QueryError(err, response, "deserializing address")
		return
	}

	state := app.GetState().(*backing.State)

	ad, exists := state.GetAccount(address, app.blockTime)
	// we use the Info field in the response to indicate whether the account exists
	response.Info = fmt.Sprintf(query.AccountInfoFmt, exists)
	ad.UpdateSettlements(app.blockTime)
	adBytes, err := ad.MarshalMsg(nil)
	if err != nil {
		app.QueryError(err, response, "serializing account data")
		return
	}

	response.Value = adBytes
}

var lastSummary query.Summary

func summaryQuery(appI interface{}, request abci.RequestQuery, response *abci.ResponseQuery) {
	app := appI.(*App)
	state := app.GetState().(*backing.State)

	// cache the last-read value for the duration of a block in case we get multiple queries
	if lastSummary.BlockHeight != app.Height() {

		var total types.Ndau
		for _, acct := range state.Accounts {
			total += acct.Balance
		}
		lastSummary.TotalNdau = total
		lastSummary.NumAccounts = len(state.Accounts)
		lastSummary.BlockHeight = app.Height()
	}

	response.Log = fmt.Sprintf("total ndau at height %d is %d, in %d accounts", lastSummary.BlockHeight, lastSummary.TotalNdau, lastSummary.NumAccounts)
	lsBytes, err := lastSummary.MarshalMsg(nil)
	if err != nil {
		app.QueryError(err, response, "serializing summary data")
		return
	}

	response.Value = lsBytes
}

func versionQuery(appI interface{}, _ abci.RequestQuery, response *abci.ResponseQuery) {
	app := appI.(*App)

	v, err := version.Get()
	if err != nil {
		app.QueryError(err, response, "getting ndaunode version")
	}
	response.Value = []byte(v)
}

func sysvarsQuery(appI interface{}, _ abci.RequestQuery, response *abci.ResponseQuery) {
	app := appI.(*App)
	names := app.systemCache.GetNames()

	sysvars := make(map[string][]byte)
	for _, n := range names {
		v := app.systemCache.GetRaw(n)
		sysvars[n] = v
	}
	buf := &bytes.Buffer{}
	err := json.NewEncoder(buf).Encode(sysvars)
	if err != nil {
		app.QueryError(err, response, "encoding sysvars")
	}
	response.Value = buf.Bytes()
}
