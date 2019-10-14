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
	"github.com/oneiro-ndev/ndau/pkg/ndauapi/routes"
	"github.com/pkg/errors"
)

// PriceInfo returns current price data for key parameters
func (c *Client) PriceInfo() (info *routes.PriceInfo, err error) {
	info = new(routes.PriceInfo)
	err = c.get(info, c.URL("price/current"))
	err = errors.Wrap(err, "fetching price info from API")
	return
}

// PriceAt returns price data at the given height
func (c *Client) PriceAt(height uint64) (info *routes.PriceInfo, err error) {
	info = new(routes.PriceInfo)
	err = c.get(info, c.URL("price/height/%d", height))
	err = errors.Wrap(err, "fetching price data from API")
	return
}
