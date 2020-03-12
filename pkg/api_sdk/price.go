package sdk

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

	srch "github.com/ndau/ndau/pkg/ndau/search"
	"github.com/ndau/ndau/pkg/ndauapi/routes"
	"github.com/pkg/errors"
)

// PriceInfo returns current price data for key parameters
func (c *Client) PriceInfo() (info *routes.PriceInfo, err error) {
	info = new(routes.PriceInfo)
	err = c.get(info, c.URL("price/current"))
	err = errors.Wrap(err, "fetching price info from API")
	return
}

// TargetPriceHistory returns historical target price data
func (c *Client) TargetPriceHistory(params srch.PriceQueryParams) ([]srch.PriceQueryResult, error) {
	return c.priceHistory(c.URL("price/target/history"), params)
}

// MarketPriceHistory returns historical market price data
func (c *Client) MarketPriceHistory(params srch.PriceQueryParams) ([]srch.PriceQueryResult, error) {
	return c.priceHistory(c.URL("price/market/history"), params)
}

func (c *Client) priceHistory(endpoint string, params srch.PriceQueryParams) ([]srch.PriceQueryResult, error) {
	var history []srch.PriceQueryResult

	// iteration proceeds while response next field is encoded params
	var pdata []byte
	var err error
	pdata, err = json.Marshal(params)
	if err != nil {
		return history, errors.Wrap(err, "marshaling initial params")
	}

	response := routes.PriceHistoryResults{
		Next: string(pdata),
	}

	for response.Next != "" {
		// unmarshal params to get next page
		err = json.Unmarshal([]byte(response.Next), &params)
		if err != nil {
			return history, errors.Wrap(err, "unmarshaling next params from api")
		}

		// perform next query
		err := c.post(params, &response, c.URL("price/target/history"))
		history = append(history, response.Items...)
		if err != nil {
			return history, errors.Wrap(err, "fetching history from api")
		}
	}

	return history, nil
}
