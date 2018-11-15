package ndau

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"net/url"

	abci "github.com/tendermint/tendermint/abci/types"

	meta "github.com/oneiro-ndev/metanode/pkg/meta/app"
	"github.com/oneiro-ndev/ndau/pkg/ndau/backing"
	srch "github.com/oneiro-ndev/ndau/pkg/ndau/search"
	"github.com/oneiro-ndev/ndau/pkg/query"
	"github.com/oneiro-ndev/ndau/pkg/version"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	"github.com/oneiro-ndev/ndaumath/pkg/types"
)

func init() {
	meta.RegisterQueryHandler(query.AccountEndpoint, accountQuery)
	meta.RegisterQueryHandler(query.SearchEndpoint, searchQuery)
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

func searchQuery(appI interface{}, request abci.RequestQuery, response *abci.ResponseQuery) {
	app := appI.(*App)

	params := string(request.GetData())
	keyValues := strings.Split(params, "&")

	// Look for params we support.
	cmd := ""
	hash := ""
	for _, keyValue := range keyValues {
		kv := strings.Split(keyValue, "=")
		if len(kv) == 2 {
			k := kv[0]
			v := kv[1]
			switch k {
			case "cmd":
				cmd = v
			case "hash":
				hash = v
			}
		}
	}

	if cmd == "" {
		app.QueryError(errors.New("Unsupported search command"),
			response, "invalid search command")
		return
	}

	search := app.GetSearch()
	if search == nil {
		app.QueryError(errors.New("Must call SetSearch()"), response, "search not available")
		return
	}

	switch cmd {
	case "heightbyblockhash":
		height, err := search.(*srch.Client).SearchBlockHash(hash)
		if err != nil {
			app.QueryError(err, response, "height by block hash search fail")
			return
		}
		value := fmt.Sprintf("%d", height)
		response.Value = []byte(value)
	case "heightbytxhash":
		txhash, err := url.PathUnescape(hash)
		if err != nil {
			app.QueryError(err, response, "height by tx hash unescape fail")
			return
		}
		height, offset, err := search.(*srch.Client).SearchTxHash(txhash)
		if err != nil {
			app.QueryError(err, response, "height by tx hash search fail")
			return
		}
		valueData := srch.TxValueData{height, offset}
		value := valueData.Marshal()
		response.Value = []byte(value)
	}

	app.QueryError(errors.New("Invalid query"), response, "invalid search params")
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
