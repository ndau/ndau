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

// GetSIB exists for compatibility; it delegates to node.PriceInfo()
func GetSIB(node *Client) (*routes.PriceInfo, error) {
	return node.PriceInfo()
}

// GetSummary exists for compatibility; it delegates to node.PriceInfo()
func GetSummary(node *Client) (*routes.PriceInfo, error) {
	return node.PriceInfo()
}

// Version delivers version information
func (c *Client) Version() (version *routes.VersionResult, err error) {
	version = new(routes.VersionResult)
	err = c.get(version, c.URL("version"))
	err = errors.Wrap(err, "getting version from API")
	return
}

// Version delivers version information
func Version(node *Client) (*routes.VersionResult, error) {
	return node.Version()
}
