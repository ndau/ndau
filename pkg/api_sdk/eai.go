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

// EAIRate returns EAI rates given certain account states
//
// The address field is just to correlate request fields with response fields;
// account data is not checked.
func (c *Client) EAIRate(query ...routes.EAIRateRequest) (response []routes.EAIRateResponse, err error) {
	response = make([]routes.EAIRateResponse, 0)
	err = c.post(query, &response, c.URL("system/eai/rate"))
	err = errors.Wrap(err, "getting EAI rates from API")
	return
}
