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
	metatx "github.com/ndau/metanode/pkg/meta/transaction"
	"github.com/ndau/ndau/pkg/ndauapi/routes"
	math "github.com/ndau/ndaumath/pkg/types"
	"github.com/pkg/errors"
)

// Prevalidate prevalidates the provided transactable
func (c *Client) Prevalidate(tx metatx.Transactable) (fee math.Ndau, sib math.Ndau, err error) {
	result := new(routes.PrevalidateResult)
	err = c.post(tx, result, c.URL("tx/prevalidate/%s", metatx.NameOf(tx)))
	if err != nil {
		err = errors.Wrap(err, "prevalidating")
		return
	}
	fee = math.Ndau(result.FeeNapu)
	sib = math.Ndau(result.SibNapu)
	return
}

// Prevalidate prevalidates the provided transactable
func Prevalidate(node *Client, tx metatx.Transactable) (fee math.Ndau, sib math.Ndau, err error) {
	return node.Prevalidate(tx)
}

// Send broadcasts and commits a transaction
func (c *Client) Send(tx metatx.Transactable) (result *routes.SubmitResult, err error) {
	result = new(routes.SubmitResult)
	err = c.post(tx, result, c.URL("tx/submit/%s", metatx.NameOf(tx)))
	err = errors.Wrap(err, "submitting")
	return
}

// SendCommit broadcasts and commits a transaction
func SendCommit(node *Client, tx metatx.Transactable) (result *routes.SubmitResult, err error) {
	return node.Send(tx)
}
