package routes

// ----- ---- --- -- -
// Copyright 2019 Oneiro NA, Inc. All Rights Reserved.
//
// Licensed under the Apache License 2.0 (the "License").  You may not use
// this file except in compliance with the License.  You can obtain a copy
// in the file LICENSE in the source distribution or at
// https://www.apache.org/licenses/LICENSE-2.0.txt
// - -- --- ---- -----

import (
	"net/http"

	"github.com/ndau/ndau/pkg/ndauapi/cfg"
	"github.com/ndau/ndau/pkg/ndauapi/reqres"
	"github.com/ndau/ndau/pkg/tool"
	"github.com/ndau/ndaumath/pkg/eai"
	"github.com/ndau/ndaumath/pkg/pricecurve"
	"github.com/ndau/ndaumath/pkg/types"
)

// PriceInfo returns data about some direct and derived quantities of ndau price information.
type PriceInfo struct {
	MarketPrice pricecurve.Nanocent `json:"marketPrice"`
	TargetPrice pricecurve.Nanocent `json:"targetPrice"`
	FloorPrice  pricecurve.Nanocent `json:"floorPrice"`
	TotalIssued types.Ndau          `json:"totalIssued"`
	TotalNdau   types.Ndau          `json:"totalNdau"`
	TotalBurned types.Ndau          `json:"totalBurned"`
	CurrentSIB  eai.Rate            `json:"sib"`
}

// getPriceInfo builds a PriceInfo object
func getPriceInfo(cf cfg.Cfg) (PriceInfo, error) {
	var oci PriceInfo

	summ, _, err := tool.GetSummary(cf.Node)
	if err != nil {
		return oci, err
	}
	oci.TotalIssued = summ.TotalIssue
	oci.TotalNdau = summ.TotalCirculation
	oci.TotalBurned = summ.TotalBurned

	sib, _, err := tool.GetSIB(cf.Node)
	if err != nil {
		return oci, err
	}

	oci.CurrentSIB = sib.SIB
	oci.TargetPrice = sib.TargetPrice
	oci.MarketPrice = sib.MarketPrice
	oci.FloorPrice = sib.FloorPrice
	return oci, err
}

// GetPriceData returns a block of price information
func GetPriceData(cf cfg.Cfg) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		response, err := getPriceInfo(cf)
		if err != nil {
			reqres.NewFromErr("price query error", err, http.StatusInternalServerError)
			return
		}
		reqres.RespondJSON(w, reqres.OKResponse(response))
	}
}

// OrderChainInfo is the old structure returned by /order/price. We only need it as long as we
// keep the deprecated function below.
type OrderChainInfo struct {
	MarketPrice float64    `json:"marketPrice"`
	TargetPrice float64    `json:"targetPrice"`
	FloorPrice  float64    `json:"floorPrice"`
	TotalIssued types.Ndau `json:"totalIssued"`
	TotalNdau   types.Ndau `json:"totalNdau"`
	SIB         float64    `json:"sib"`
	PriceUnits  string     `json:"priceUnit"`
}

// GetOrderData returns a block of price information, but converts it to a format that is compatible
// with the old /order/current API. This is for use while v1.8 of the wallet is still extant.
func GetOrderData(cf cfg.Cfg) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		pricedata, err := getPriceInfo(cf)
		if err != nil {
			reqres.NewFromErr("price query error", err, http.StatusInternalServerError)
			return
		}
		oci := OrderChainInfo{
			// converting from nanocents to floating point dollars means multiplying by 10^11
			MarketPrice: float64(pricedata.MarketPrice) / 100000000000.0,
			TargetPrice: float64(pricedata.TargetPrice) / 100000000000.0,
			FloorPrice:  2.40,
			TotalIssued: pricedata.TotalIssued,
			TotalNdau:   pricedata.TotalNdau,
			SIB:         1,
			PriceUnits:  "USD",
		}

		reqres.RespondJSON(w, reqres.OKResponse(oci))
	}
}
