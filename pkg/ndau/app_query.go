package ndau

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"

	meta "github.com/oneiro-ndev/metanode/pkg/meta/app"
	metasrch "github.com/oneiro-ndev/metanode/pkg/meta/search"
	metatx "github.com/oneiro-ndev/metanode/pkg/meta/transaction"
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
	meta.RegisterQueryHandler(query.DelegatesEndpoint, delegatesQuery)
	meta.RegisterQueryHandler(query.SysvarHistoryEndpoint, sysvarHistoryQuery)
	meta.RegisterQueryHandler(query.PrevalidateEndpoint, prevalidateQuery)
	meta.RegisterQueryHandler(query.SearchEndpoint, searchQuery)
	meta.RegisterQueryHandler(query.SIBEndpoint, sibQuery)
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

	ad, exists := app.getAccount(address)
	// we use the Info field in the response to indicate whether the account exists
	response.Info = fmt.Sprintf(query.AccountInfoFmt, exists)
	ad.UpdateSettlements(app.BlockTime())
	// update the WAA field to get up-to-the-microsecond values
	ad.WeightedAverageAge += app.BlockTime().Since(ad.LastWAAUpdate)
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

	var params srch.AccountHistoryParams
	err := json.Unmarshal(request.GetData(), &params)
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

	var params srch.AccountHistoryParams
	err := json.Unmarshal(request.GetData(), &params)
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

	retval := query.AccountListQueryResponse{
		NumAccounts: len(names),
		FirstIndex:  start,
		PageSize:    params.PageSize,
		PageIndex:   params.PageIndex,
		Accounts:    names[start:end],
	}
	rBytes, err := retval.MarshalMsg(nil)
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

	mtx, err := metatx.Unmarshal(request.GetData(), TxIDs)
	if err != nil {
		app.QueryError(err, response, "deserializing transactable")
		return
	}

	tx, ok := mtx.(NTransactable)
	if !ok {
		app.QueryError(
			fmt.Errorf("tx %s not an NTransactable", metatx.NameOf(mtx)),
			response,
			"converting metatx.Transactable to NTransactable",
		)
		return
	}

	// we use the Info field to communicate the estimated tx fee
	fee, err := app.calculateTxFee(tx)
	if err != nil {
		app.QueryError(err, response, "calculating tx fee")
		return
	}

	sib, err := app.calculateSIB(tx)
	if err != nil {
		app.QueryError(err, response, "calculating sib")
		return
	}

	response.Info = fmt.Sprintf(query.PrevalidateInfoFmt, fee, sib)

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

	var params srch.QueryParams
	err := json.Unmarshal(request.GetData(), &params)
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
		lastSummary.TotalRFE = state.TotalRFE
		lastSummary.TotalIssue = state.TotalIssue
		lastSummary.TotalBurned = state.TotalBurned
		// the total ndau in circulation is the total in all accounts, excluding
		// the amount of ndau that have been released but not issued
		lastSummary.TotalCirculation = lastSummary.TotalNdau - ((lastSummary.TotalRFE - lastSummary.TotalIssue) + lastSummary.TotalBurned)
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

func sysvarsQuery(appI interface{}, request abci.RequestQuery, response *abci.ResponseQuery) {
	app := appI.(*App)

	// decode request
	var err error
	var svr query.SysvarsRequest
	if len(request.Data) > 0 {
		_, err = svr.UnmarshalMsg(request.Data)
		if err != nil {
			app.QueryError(err, response, "decoding sysvars request")
			return
		}
	}

	// get sysvars
	sv := app.GetState().(*backing.State).Sysvars

	// apply filter as required
	svo := make(map[string][]byte)
	filter := []string(svr)
	if len(filter) > 0 {
		for _, f := range filter {
			svo[f] = sv[f]
		}
	} else {
		svo = sv
	}

	// return
	resp := query.SysvarsResponse(svo)
	response.Value, err = resp.MarshalMsg(nil)
	if err != nil {
		app.QueryError(err, response, "encoding sysvars response")
	}
}

func sysvarHistoryQuery(
	appI interface{}, request abci.RequestQuery, response *abci.ResponseQuery,
) {
	app := appI.(*App)

	search := app.GetSearch()
	if search == nil {
		app.QueryError(errors.New("Must call SetSearch()"), response, "search not available")
		return
	}
	client := search.(*srch.Client)

	var params srch.SysvarHistoryParams
	err := json.Unmarshal(request.GetData(), &params)
	if err != nil {
		app.QueryError(
			errors.New("Cannot decode search params json"), response, "invalid search query")
		return
	}

	khr, err := client.SearchSysvarHistory(params.Name, params.PageIndex, params.PageSize)
	if err != nil {
		app.QueryError(err, response, "sysvar history search fail")
		return
	}

	khBytes, err := khr.MarshalMsg(nil)
	response.Value = khBytes
	app.QueryError(err, response, "sysvar history byte serialization fail")
}

func delegatesQuery(appI interface{}, _ abci.RequestQuery, response *abci.ResponseQuery) {
	app := appI.(*App)
	state := app.GetState().(*backing.State)

	dr := make(query.DelegatesResponse, 0, len(state.Delegates))
	for nodeS, acctsS := range state.Delegates {
		node, err := address.Validate(nodeS)
		if err != nil {
			response.Info += fmt.Sprintf("bad node address: %q\n", nodeS)
			continue
		}
		delegate := query.DelegateList{
			Node: node,
		}

		for acctS := range acctsS {
			acct, err := address.Validate(acctS)
			if err != nil {
				response.Info += fmt.Sprintf("bad acct address: %q\n", acctS)
				response.Info += fmt.Sprintf("        for node: %s\n", node)
				continue
			}
			delegate.Delegated = append(delegate.Delegated, acct)
		}

		dr = append(dr, delegate)
	}

	bytes, err := dr.MarshalMsg(nil)
	if err != nil {
		app.QueryError(err, response, "failed to marshal delegates response")
		return
	}

	response.Value = bytes
}

func sibQuery(appI interface{}, request abci.RequestQuery, response *abci.ResponseQuery) {
	var err error
	app := appI.(*App)
	state := app.GetState().(*backing.State)
	resp := query.SIBResponse{
		SIB:         state.SIB,
		MarketPrice: state.MarketPrice,
		TargetPrice: state.TargetPrice,
	}
	response.Info = resp.SIB.String()
	response.Value, err = resp.MarshalMsg(nil)
	if err != nil {
		app.QueryError(err, response, "encoding SIB")
	}
}
