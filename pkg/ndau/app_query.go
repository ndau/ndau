package ndau

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"

	meta "github.com/oneiro-ndev/metanode/pkg/meta/app"
	metasrch "github.com/oneiro-ndev/metanode/pkg/meta/search"
	"github.com/oneiro-ndev/metanode/pkg/meta/transaction"
	"github.com/oneiro-ndev/ndau/pkg/ndau/backing"
	srch "github.com/oneiro-ndev/ndau/pkg/ndau/search"
	"github.com/oneiro-ndev/ndau/pkg/query"
	"github.com/oneiro-ndev/ndau/pkg/version"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	"github.com/oneiro-ndev/ndaumath/pkg/types"
	abci "github.com/tendermint/tendermint/abci/types"
)

func init() {
	meta.RegisterQueryHandler(query.AccountEndpoint, accountQuery)
	meta.RegisterQueryHandler(query.AccountHistoryEndpoint, accountHistoryQuery)
	meta.RegisterQueryHandler(query.AccountListEndpoint, accountListQuery)
	meta.RegisterQueryHandler(query.DateRangeEndpoint, dateRangeQuery)
	meta.RegisterQueryHandler(query.PrevalidateEndpoint, prevalidateQuery)
	meta.RegisterQueryHandler(query.SearchEndpoint, searchQuery)
	meta.RegisterQueryHandler(query.SidechainTxExistsEndpoint, sidechainTxExistsQuery)
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

func accountHistoryQuery(
	appI interface{}, request abci.RequestQuery, response *abci.ResponseQuery,
) {
	app := appI.(*App)

	search := app.GetSearch()
	if search == nil {
		app.QueryError(errors.New("Must call SetSearch()"), response, "search not available")
		return
	}
	client := search.(*srch.Client)

	paramsString := string(request.GetData())
	var params srch.AccountHistoryParams
	err := json.NewDecoder(strings.NewReader(paramsString)).Decode(&params)
	if err != nil {
		app.QueryError(
			errors.New("Cannot decode search params json"), response, "invalid search query")
		return
	}

	// The address was already validated by the caller.
	ahr, err := client.SearchAccountHistory(params.Address, params.PageIndex, params.PageSize)
	if err != nil {
		app.QueryError(err, response, "account history search fail")
		return
	}

	ahBytes := []byte(ahr.Marshal())
	response.Value = ahBytes
}

func accountListQuery(appI interface{}, request abci.RequestQuery, response *abci.ResponseQuery) {
	app := appI.(*App)

	paramsString := string(request.GetData())
	var params srch.AccountHistoryParams
	err := json.NewDecoder(strings.NewReader(paramsString)).Decode(&params)
	if err != nil {
		app.QueryError(
			errors.New("cannot decode search params json"), response, "invalid search query")
		return
	}

	state := app.GetState().(*backing.State)
	// we need to get all the names so we can sort them
	names := make([]string, len(state.Accounts))
	ix := 0
	for k := range state.Accounts {
		names[ix] = k
		ix++
	}
	sort.Sort(sort.StringSlice(names))
	start := params.PageIndex * params.PageSize
	if start > len(names) {
		start = len(names)
	}
	end := (params.PageIndex + 1) * params.PageSize
	if end > len(names) {
		end = len(names)
	}

	retval := names[start:end]
	rBytes, err := json.Marshal(retval)
	if err != nil {
		app.QueryError(err, response, "serializing account data")
		return
	}

	response.Value = rBytes
}

func dateRangeQuery(appI interface{}, request abci.RequestQuery, response *abci.ResponseQuery) {
	app := appI.(*App)

	search := app.GetSearch()
	if search == nil {
		app.QueryError(errors.New("Must call SetSearch()"), response, "search not available")
		return
	}
	client := search.(*srch.Client)

	paramsString := string(request.GetData())
	var req metasrch.DateRangeRequest
	req.Unmarshal(paramsString)

	firstHeight, lastHeight, err :=
		client.Client.SearchDateRange(req.FirstTimestamp, req.LastTimestamp)
	if err != nil {
		app.QueryError(err, response, "date range search fail")
		return
	}

	result := metasrch.DateRangeResult{FirstHeight: firstHeight, LastHeight: lastHeight}
	ahBytes := []byte(result.Marshal())
	response.Value = ahBytes
}

func prevalidateQuery(appI interface{}, request abci.RequestQuery, response *abci.ResponseQuery) {
	app := appI.(*App)

	tx, err := metatx.Unmarshal(request.GetData(), TxIDs)
	if err != nil {
		app.QueryError(err, response, "deserializing transactable")
		return
	}

	// we use the Info field to communicate the estimated tx fee
	fee, err := app.calculateTxFee(tx)
	if err != nil {
		app.QueryError(err, response, "calculating tx fee")
		return
	}
	response.Info = fmt.Sprintf(query.PrevalidateInfoFmt, fee)

	err = tx.Validate(appI)
	if err != nil {
		app.QueryError(err, response, "validating transactable")
		return
	}

	return
}

func searchQuery(appI interface{}, request abci.RequestQuery, response *abci.ResponseQuery) {
	app := appI.(*App)

	search := app.GetSearch()
	if search == nil {
		app.QueryError(errors.New("Must call SetSearch()"), response, "search not available")
		return
	}
	client := search.(*srch.Client)

	paramsString := string(request.GetData())
	var params srch.QueryParams
	err := json.NewDecoder(strings.NewReader(paramsString)).Decode(&params)
	if err != nil {
		app.QueryError(
			errors.New("Cannot decode search params json"), response, "invalid search query")
		return
	}

	switch params.Command {
	case srch.HeightByBlockHashCommand:
		height, err := client.SearchBlockHash(params.Hash)
		if err != nil {
			app.QueryError(err, response, "height by block hash search fail")
			return
		}
		value := fmt.Sprintf("%d", height)
		response.Value = []byte(value)
	case srch.HeightByTxHashCommand:
		height, offset, err := client.SearchTxHash(params.Hash)
		if err != nil {
			app.QueryError(err, response, "height by tx hash search fail")
			return
		}
		valueData := srch.TxValueData{BlockHeight: height, TxOffset: offset}
		value := valueData.Marshal()
		response.Value = []byte(value)
	default:
		app.QueryError(errors.New("Invalid query"), response, "invalid search params")
	}
}

func sidechainTxExistsQuery(appI interface{}, request abci.RequestQuery, response *abci.ResponseQuery) {
	app := appI.(*App)

	stxq := new(query.SidechainTxExistsQuery)
	_, err := stxq.UnmarshalMsg(request.GetData())
	if err != nil {
		app.QueryError(err, response, "unmarshalling SidechainTxExistsQuery")
		return
	}

	acct, _ := app.GetState().(*backing.State).GetAccount(stxq.Source, app.blockTime)
	key := sidechainPayment(stxq.SidechainID, stxq.TxHash)

	_, exists := acct.SidechainPayments[key]

	response.Info = fmt.Sprintf(query.SidechainTxExistsInfoFmt, exists)
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
