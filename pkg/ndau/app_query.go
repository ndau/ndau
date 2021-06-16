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
	"encoding/json"
	"errors"
	"fmt"
	"sort"

	meta "github.com/ndau/metanode/pkg/meta/app"
	metasrch "github.com/ndau/metanode/pkg/meta/search"
	metatx "github.com/ndau/metanode/pkg/meta/transaction"
	"github.com/ndau/ndau/pkg/ndau/backing"
	srch "github.com/ndau/ndau/pkg/ndau/search"
	"github.com/ndau/ndau/pkg/query"
	"github.com/ndau/ndau/pkg/version"
	"github.com/ndau/ndaumath/pkg/address"
	"github.com/ndau/ndaumath/pkg/types"
	abci "github.com/tendermint/tendermint/abci/types"
)

func init() {
	meta.RegisterQueryHandler(query.AccountEndpoint, accountQuery)
	meta.RegisterQueryHandler(query.AccountHistoryEndpoint, accountHistoryQuery)
	meta.RegisterQueryHandler(query.AccountListEndpoint, accountListQuery)
	meta.RegisterQueryHandler(query.DateRangeEndpoint, dateRangeQuery)
	meta.RegisterQueryHandler(query.DelegatesEndpoint, delegatesQuery)
	meta.RegisterQueryHandler(query.NodesEndpoint, nodesQuery)
	meta.RegisterQueryHandler(query.PrevalidateEndpoint, prevalidateQuery)
	meta.RegisterQueryHandler(query.PriceMarketEndpoint, priceQuery)
	meta.RegisterQueryHandler(query.PriceTargetEndpoint, priceQuery)
	meta.RegisterQueryHandler(query.SearchEndpoint, searchQuery)
	meta.RegisterQueryHandler(query.SIBEndpoint, sibQuery)
	meta.RegisterQueryHandler(query.SummaryEndpoint, summaryQuery)
	meta.RegisterQueryHandler(query.SysvarHistoryEndpoint, sysvarHistoryQuery)
	meta.RegisterQueryHandler(query.SysvarsEndpoint, sysvarsQuery)
	meta.RegisterQueryHandler(query.VersionEndpoint, versionQuery)
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
	ad.UpdateRecourses(app.BlockTime())
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
		app.QueryError(errors.New("must call SetSearch()"), response, "search not available")
		return
	}
	client := search.(*srch.Client)

	var params srch.AccountHistoryParams
	err := json.Unmarshal(request.GetData(), &params)
	if err != nil {
		app.QueryError(
			errors.New("cannot decode search params json"), response, "invalid search query")
		return
	}

	// The address was already validated by the caller.
	ahr, err := client.SearchAccountHistory(params.Address, params.AfterHeight, params.Limit)
	if err != nil {
		app.QueryError(err, response, "account history search fail")
		return
	}

	ahBytes := []byte(ahr.Marshal())
	response.Value = ahBytes
}

func accountListQuery(appI interface{}, request abci.RequestQuery, response *abci.ResponseQuery) {
	app := appI.(*App)

	var params srch.AccountListParams
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
	numAccounts := len(names)
	sort.Sort(sort.StringSlice(names))
	// Reduce the full results list down to the requested portion.  There is some wasted effort with
	// this approach, but we support the worst case, which is to return all results.  In practice,
	// getting the full list from the underlying index is fast, with tolerable sorting speed.
	offsetStart := sort.Search(len(names), func(n int) bool {
		return names[n] > params.After
	})
	names = names[offsetStart:]
	// only specify nextafter if there are more things to query, which happens when we have
	// to truncate the end of the list
	nextafter := ""
	if params.Limit > 0 && len(names) > params.Limit {
		names = names[:params.Limit]
		nextafter = names[len(names)-1]
	}

	retval := query.AccountListQueryResponse{
		NumAccounts: numAccounts,
		FirstIndex:  offsetStart,
		After:       params.After,
		NextAfter:   nextafter,
		Accounts:    names,
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
		app.QueryError(errors.New("must call SetSearch()"), response, "search not available")
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
}

func searchQuery(appI interface{}, request abci.RequestQuery, response *abci.ResponseQuery) {
	app := appI.(*App)

	search := app.GetSearch()
	if search == nil {
		app.QueryError(errors.New("must call SetSearch()"), response, "search not available")
		return
	}
	client := search.(*srch.Client)

	var params srch.QueryParams
	err := json.Unmarshal(request.GetData(), &params)
	if err != nil {
		app.QueryError(
			errors.New("cannot decode search params json"), response, "invalid search query")
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
		valueData, err := client.SearchTxHash(params.Hash)
		if err != nil {
			app.QueryError(err, response, "height by tx hash search fail")
			return
		}
		value := valueData.Marshal()
		response.Value = []byte(value)
	case srch.HeightsByTxTypesCommand:
		valueData, err := client.SearchTxTypes(params.Hash, params.Types, params.Limit)
		if err != nil {
			app.QueryError(err, response, "heights by tx types search fail")
			return
		}
		value := valueData.Marshal()
		response.Value = []byte(value)
	default:
		app.QueryError(errors.New("invalid query"), response, "invalid search params")
	}
}

var lastSummary query.Summary

func getLastSummary(app *App) query.Summary {
	// cache the last-read value for the duration of a block in case we get multiple queries
	if lastSummary.BlockHeight != app.Height() {
		state := app.GetState().(*backing.State)

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

		// Tracking TotalNdau and TotalCirculation together is a bad idea, especially when queries
		// return TotalCirculation when TotalNdau is requested. But cleaning that up makes the
		// implementation of a height gate (which is necessary) very messy, so I'm leaving it.

		// TotalBurned should never have been subtracted from TotalNdau - that's an old bug that's being fixed. But it's inside
		// the height gate because the calculation of the floor price - and, therefore, SIB - depends on it. And this calculation
		// changes the rules for RFE. All ndau released from the endowment are immediately in circulation. The TotalIssued value
		// is only used to calculate the current target price.

		if app.IsFeatureActive("AllRFEInCirculation") {
			lastSummary.TotalCirculation = lastSummary.TotalNdau
		} else {
			// the total ndau in circulation is the total in all accounts, excluding
			// the amount of ndau that have been released but not issued
			lastSummary.TotalCirculation = lastSummary.TotalNdau - ((lastSummary.TotalRFE - lastSummary.TotalIssue) + lastSummary.TotalBurned)
		}
	}
	return lastSummary
}

func summaryQuery(appI interface{}, request abci.RequestQuery, response *abci.ResponseQuery) {
	app := appI.(*App)

	ls := getLastSummary(app)
	response.Log = fmt.Sprintf("total ndau at height %d is %d, in %d accounts", ls.BlockHeight, ls.TotalNdau, ls.NumAccounts)

	lsBytes, err := ls.MarshalMsg(nil)
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
		app.QueryError(errors.New("must call SetSearch()"), response, "search not available")
		return
	}
	client := search.(*srch.Client)

	var params srch.SysvarHistoryParams
	err := json.Unmarshal(request.GetData(), &params)
	if err != nil {
		app.QueryError(
			errors.New("cannot decode search params json"), response, "invalid search query")
		return
	}

	khr, err := client.SearchSysvarHistory(params.Name, params.AfterHeight, params.Limit)
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
	fp, err := floorPrice(app, state.GetEndowmentNAV())
	if err != nil {
		app.QueryError(err, response, "calculating floor price")
		return
	}
	resp := query.SIBResponse{
		SIB:         state.SIB,
		MarketPrice: state.MarketPrice,
		TargetPrice: state.TargetPrice,
		FloorPrice:  fp,
	}
	response.Info = resp.SIB.String()
	response.Value, err = resp.MarshalMsg(nil)
	if err != nil {
		app.QueryError(err, response, "encoding SIB")
	}
}

func nodesQuery(appI interface{}, request abci.RequestQuery, response *abci.ResponseQuery) {
	app := appI.(*App)
	state := app.GetState().(*backing.State)

	nresp := make(query.NodesResponse)
	for addr, node := range state.Nodes {
		nresp[addr] = query.NodeExtra{
			Node:         node,
			Registration: node.GetRegistration(),
		}
	}

	var err error
	response.Value, err = nresp.MarshalMsg(nil)
	if err != nil {
		app.QueryError(err, response, "marshaling node response")
	}
}

func priceQuery(appI interface{}, request abci.RequestQuery, response *abci.ResponseQuery) {
	app := appI.(*App)
	search := app.GetSearch().(*srch.Client)

	// chose the appropriate search function
	var sf func(params srch.PriceQueryParams) (srch.PriceQueryResults, error)
	switch request.Path {
	case query.PriceMarketEndpoint:
		sf = search.SearchMarketPrice
	case query.PriceTargetEndpoint:
		sf = search.SearchTargetPrice
	}

	// unpack params
	var pqp srch.PriceQueryParams
	if len(request.Data) > 0 {
		_, err := pqp.After.Timestamp.UnmarshalMsg(request.Data)
		if err != nil {
			app.QueryError(err, response, "unmarshaling query params")
			return
		}
	}

	// perform search
	pqr, err := sf(pqp)
	if err != nil {
		app.QueryError(err, response, "searching for price data")
		return
	}

	// pack response
	response.Value, err = pqr.MarshalMsg(nil)
	if err != nil {
		app.QueryError(err, response, "marshaling price data results")
	}
}
